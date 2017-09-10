package nut

import (
	"context"
	"crypto/x509/pkix"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/garyburd/redigo/redis"
	"github.com/google/uuid"
	"github.com/ikeikeikeike/go-sitemap-generator/stm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/steinbacher/goose"
	"github.com/urfave/cli"
	"golang.org/x/text/language"
	"golang.org/x/tools/blog/atom"
)

const (
	postgresqlDriver = "postgres"
)

func init() {
	host, _ := os.Hostname()

	AddConsoleTask(
		cli.Command{
			Name:    "users",
			Aliases: []string{"us"},
			Usage:   "users operations",
			Subcommands: []cli.Command{
				{
					Name:    "list",
					Aliases: []string{"l"},
					Usage:   "list users",
					Action: Open(func(*cli.Context) error {

						var users []User
						if err := DB().Select([]string{"name", "email", "uid"}).
							Find(&users).Error; err != nil {
							return err
						}
						fmt.Printf("UID\t\t\t\t\tFULL-NAME<EMAIL>\n")
						for _, u := range users {
							fmt.Printf("%s\t%s<%s>\n", u.UID, u.Name, u.Email)
						}
						return nil
					}),
				},
				{
					Name:    "role",
					Aliases: []string{"r"},
					Usage:   "apply/deny role to user",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name, n",
							Value: "",
							Usage: "role's name",
						},
						cli.StringFlag{
							Name:  "user, u",
							Value: "",
							Usage: "user's uid",
						},
						cli.IntFlag{
							Name:  "years, y",
							Value: 10,
							Usage: "years",
						},
						cli.BoolFlag{
							Name:  "deny, d",
							Usage: "deny mode",
						},
					},
					Action: Open(func(c *cli.Context) error {
						uid := c.String("user")
						name := c.String("name")
						deny := c.Bool("deny")
						years := c.Int("years")
						if uid == "" || name == "" {
							cli.ShowSubcommandHelp(c)
							return nil
						}

						user, err := GetUserByUID(uid)
						if err != nil {
							return err
						}

						role, err := GetRole(name, DefaultResourceType, DefaultResourceID)
						if err != nil {
							return err
						}
						if deny {
							return Deny(role.ID, user.ID)
						}
						return Allow(role.ID, user.ID, years, 0, 0)
					}),
				},
			},
		},
		cli.Command{
			Name:    "server",
			Aliases: []string{"s"},
			Usage:   "start the app server",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "worker, w",
					Usage: "with a worker",
				},
			},
			Action: Open(func(c *cli.Context) error {
				// ---------------
				if c.Bool("worker") {
					name, _ := os.Hostname()
					go func() {
						for {
							if err := Receive(name); err != nil {
								log.Error(err)
							}
						}
					}()
				}
				// http
				rt := Router()
				addr := fmt.Sprintf(":%d", viper.GetInt("server.port"))
				if !IsProduction() {
					return rt.Run(addr)
				}
				// gracefully
				srv := &http.Server{
					Addr:    addr,
					Handler: rt,
				}

				go func() {
					// service connections
					if err := srv.ListenAndServe(); err != nil {
						log.Error(err)
					}
				}()

				// Wait for interrupt signal to gracefully shutdown the server with
				// a timeout of 5 seconds.
				quit := make(chan os.Signal)
				signal.Notify(quit, os.Interrupt)
				<-quit
				log.Println("Shutdown Server ...")

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := srv.Shutdown(ctx); err != nil {
					log.Fatal("Server Shutdown:", err)
				}
				log.Println("Server exist")
				return nil
			}),
		},
		cli.Command{
			Name:  "seo",
			Usage: "generate sitemap.xml.gz/rss.atom/robots.txt ...etc",
			Action: Open(func(*cli.Context) error {
				root := "public"
				os.MkdirAll(root, 0755)
				if err := writeSitemap(root); err != nil {
					return err
				}

				for _, lang := range Languages() {
					if err := writeRssAtom(root, lang); err != nil {
						return err
					}
				}
				if err := writeRobotsTxt(root); err != nil {
					return err
				}
				if err := writeGoogleVerify(root); err != nil {
					return err
				}
				if err := writeBaiduVerify(root); err != nil {
					return err
				}
				return nil
			}),
		},
		cli.Command{
			Name:    "worker",
			Aliases: []string{"w"},
			Usage:   "start the worker progress",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: host,
					Usage: "worker's name",
				},
			},
			Action: Open(func(c *cli.Context) error {
				name := c.String("name")
				if name == "" {
					cli.ShowSubcommandHelp(c)
					return nil
				}
				return Receive(name)
			}),
		},
		cli.Command{
			Name:    "redis",
			Aliases: []string{"re"},
			Usage:   "open redis connection",
			Action: func(c *cli.Context) error {
				if err := viper.ReadInConfig(); err != nil {
					return err
				}
				return Shell(
					"redis-cli",
					"-h", viper.GetString("redis.host"),
					"-p", viper.GetString("redis.port"),
					"-n", viper.GetString("redis.db"),
				)
			},
		},
		cli.Command{
			Name:    "cache",
			Aliases: []string{"c"},
			Usage:   "cache operations",
			Subcommands: []cli.Command{
				{
					Name:    "list",
					Usage:   "list all cache keys",
					Aliases: []string{"l"},
					Action: Open(func(_ *cli.Context) error {
						c := Redis().Get()
						defer c.Close()
						keys, err := redis.Strings(c.Do("KEYS", "*"))
						if err != nil {
							return err
						}
						for _, k := range keys {
							fmt.Println(k)
						}
						return nil
					}),
				},
				{
					Name:    "clear",
					Usage:   "clear cache items",
					Aliases: []string{"c"},
					Action: Open(func(_ *cli.Context) error {
						c := Redis().Get()
						defer c.Close()
						keys, err := redis.Values(c.Do("KEYS", "*"))
						if err == nil && len(keys) > 0 {
							_, err = c.Do("DEL", keys...)
						}
						return err
					}),
				},
			},
		},
		cli.Command{
			Name:    "database",
			Aliases: []string{"db"},
			Usage:   "database operations",
			Subcommands: []cli.Command{
				{
					Name:    "example",
					Usage:   "scripts example for create database and user",
					Aliases: []string{"e"},
					Action: func(c *cli.Context) error {
						if err := viper.ReadInConfig(); err != nil {
							return err
						}
						drv := viper.GetString("database.driver")
						args := viper.GetStringMapString("database.args")
						var err error
						switch drv {
						case postgresqlDriver:
							fmt.Printf("CREATE USER %s WITH PASSWORD '%s';\n", args["user"], args["password"])
							fmt.Printf("CREATE DATABASE %s WITH ENCODING='UTF8';\n", args["dbname"])
							fmt.Printf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s;\n", args["dbname"], args["user"])
						default:
							err = fmt.Errorf("unknown driver %s", drv)
						}
						return err
					},
				},
				{
					Name:    "migrate",
					Usage:   "migrate the DB to the most recent version available",
					Aliases: []string{"m"},
					Action: func(c *cli.Context) error {
						if err := viper.ReadInConfig(); err != nil {
							return err
						}
						conf, err := dbConf()
						if err != nil {
							return err
						}

						target, err := goose.GetMostRecentDBVersion(conf.MigrationsDir)
						if err != nil {
							return err
						}

						return goose.RunMigrations(conf, conf.MigrationsDir, target)
					},
				},
				{
					Name:    "rollback",
					Usage:   "roll back the version by 1",
					Aliases: []string{"r"},
					Action: func(c *cli.Context) error {
						if err := viper.ReadInConfig(); err != nil {
							return err
						}
						conf, err := dbConf()
						if err != nil {
							return err
						}

						current, err := goose.GetDBVersion(conf)
						if err != nil {
							return err
						}

						previous, err := goose.GetPreviousDBVersion(conf.MigrationsDir, current)
						if err != nil {
							return err
						}

						return goose.RunMigrations(conf, conf.MigrationsDir, previous)
					},
				},
				{
					Name:    "version",
					Usage:   "dump the migration status for the current DB",
					Aliases: []string{"v"},
					Action: func(c *cli.Context) error {
						if err := viper.ReadInConfig(); err != nil {
							return err
						}
						conf, err := dbConf()
						if err != nil {
							return err
						}

						// collect all migrations
						migrations, err := goose.CollectMigrations(conf.MigrationsDir)
						if err != nil {
							return err
						}

						db, err := goose.OpenDBFromDBConf(conf)
						if err != nil {
							return err
						}
						defer db.Close()

						// must ensure that the version table exists if we're running on a pristine DB
						if _, err = goose.EnsureDBVersion(conf, db); err != nil {
							return err
						}

						fmt.Println("    Applied At                  Migration")
						fmt.Println("    =======================================")
						for _, m := range migrations {
							if err = printMigrationStatus(db, m.Version, filepath.Base(m.Source)); err != nil {
								return err
							}
						}
						return nil
					},
				},
				{
					Name:    "connect",
					Usage:   "connect database",
					Aliases: []string{"c"},
					Action: func(c *cli.Context) error {
						if err := viper.ReadInConfig(); err != nil {
							return err
						}
						drv := viper.GetString("database.driver")
						args := viper.GetStringMapString("database.args")
						var err error
						switch drv {
						case postgresqlDriver:
							err = Shell("psql",
								"-h", args["host"],
								"-p", args["port"],
								"-U", args["user"],
								args["dbname"],
							)
						default:
							err = fmt.Errorf("unknown driver %s", drv)
						}
						return err
					},
				},
				{
					Name:    "create",
					Usage:   "create database",
					Aliases: []string{"n"},
					Action: func(c *cli.Context) error {
						if err := viper.ReadInConfig(); err != nil {
							return err
						}
						drv := viper.GetString("database.driver")
						args := viper.GetStringMapString("database.args")
						var err error
						switch drv {
						case postgresqlDriver:
							err = Shell("psql",
								"-h", args["host"],
								"-p", args["port"],
								"-U", "postgres",
								"-c", fmt.Sprintf(
									"CREATE DATABASE %s WITH ENCODING='UTF8'",
									args["dbname"],
								),
							)
						default:
							err = fmt.Errorf("unknown driver %s", drv)
						}
						return err
					},
				},
				{
					Name:    "drop",
					Usage:   "drop database",
					Aliases: []string{"d"},
					Action: func(c *cli.Context) error {
						if err := viper.ReadInConfig(); err != nil {
							return err
						}
						drv := viper.GetString("database.driver")
						args := viper.GetStringMapString("database.args")
						var err error
						switch drv {
						case postgresqlDriver:
							err = Shell("psql",
								"-h", args["host"],
								"-p", args["port"],
								"-U", "postgres",
								"-c", fmt.Sprintf("DROP DATABASE %s", args["dbname"]),
							)
						default:
							err = fmt.Errorf("unknown driver %s", drv)
						}
						return err
					},
				},
			},
		},
		cli.Command{
			Name:    "generate",
			Aliases: []string{"g"},
			Usage:   "generate file template",
			Subcommands: []cli.Command{
				{
					Name:    "config",
					Aliases: []string{"c"},
					Usage:   "generate config file",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "environment, e",
							Value: "development",
							Usage: "environment, like: development, production, stage, test...",
						},
					},
					Action: func(c *cli.Context) error {
						const fn = "config.toml"
						if _, err := os.Stat(fn); err == nil {
							return fmt.Errorf("file %s already exists", fn)
						}
						fmt.Printf("generate file %s\n", fn)

						viper.Set("env", c.String("environment"))
						args := viper.AllSettings()
						fd, err := os.Create(fn)
						if err != nil {
							return err
						}
						defer fd.Close()
						end := toml.NewEncoder(fd)
						return end.Encode(args)
					},
				},
				{
					Name:    "nginx",
					Aliases: []string{"ng"},
					Usage:   "generate nginx.conf",
					Action: func(c *cli.Context) error {
						if err := viper.ReadInConfig(); err != nil {
							return err
						}

						pwd, err := os.Getwd()
						if err != nil {
							return err
						}

						name := viper.GetString("server.name")
						fn := path.Join("etc", "nginx", "sites-enabled", name+".conf")
						if err = os.MkdirAll(path.Dir(fn), 0700); err != nil {
							return err
						}
						fmt.Printf("generate file %s\n", fn)
						fd, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
						if err != nil {
							return err
						}
						defer fd.Close()

						tpl, err := template.ParseFiles(path.Join("templates", "nginx.conf"))
						if err != nil {
							return err
						}

						return tpl.Execute(fd, struct {
							Port    int
							Root    string
							Name    string
							Ssl     bool
							Version string
						}{
							Name:    name,
							Port:    viper.GetInt("server.port"),
							Root:    pwd,
							Ssl:     viper.GetBool("server.ssl"),
							Version: "v1",
						})
					},
				},
				{
					Name:    "openssl",
					Aliases: []string{"ssl"},
					Usage:   "generate ssl certificates",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name, n",
							Usage: "name",
						},
						cli.StringFlag{
							Name:  "country, c",
							Value: "Earth",
							Usage: "country",
						},
						cli.StringFlag{
							Name:  "organization, o",
							Value: "Mother Nature",
							Usage: "organization",
						},
						cli.IntFlag{
							Name:  "years, y",
							Value: 1,
							Usage: "years",
						},
					},
					Action: func(c *cli.Context) error {
						name := c.String("name")
						if len(name) == 0 {
							cli.ShowCommandHelp(c, "openssl")
							return nil
						}
						root := path.Join("etc", "ssl", name)

						key, crt, err := CreateCertificate(
							true,
							pkix.Name{
								Country:      []string{c.String("country")},
								Organization: []string{c.String("organization")},
							},
							c.Int("years"),
						)
						if err != nil {
							return err
						}

						fnk := path.Join(root, "key.pem")
						fnc := path.Join(root, "crt.pem")

						fmt.Printf("generate pem file %s\n", fnk)
						err = WritePemFile(fnk, "RSA PRIVATE KEY", key, 0600)
						fmt.Printf("test: openssl rsa -noout -text -in %s\n", fnk)

						if err == nil {
							fmt.Printf("generate pem file %s\n", fnc)
							err = WritePemFile(fnc, "CERTIFICATE", crt, 0444)
							fmt.Printf("test: openssl x509 -noout -text -in %s\n", fnc)
						}
						if err == nil {
							fmt.Printf(
								"verify: diff <(openssl rsa -noout -modulus -in %s) <(openssl x509 -noout -modulus -in %s)",
								fnk,
								fnc,
							)
						}
						fmt.Println()
						return err
					},
				},
				{
					Name:    "migration",
					Usage:   "generate migration file",
					Aliases: []string{"m"},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name, n",
							Usage: "name",
						},
					},
					Action: func(c *cli.Context) error {
						if err := viper.ReadInConfig(); err != nil {
							return err
						}
						name := c.String("name")
						if len(name) == 0 {
							cli.ShowCommandHelp(c, "migration")
							return nil
						}
						cfg, err := dbConf()
						if err != nil {
							return err
						}
						if err = os.MkdirAll(cfg.MigrationsDir, 0700); err != nil {
							return err
						}
						file, err := goose.CreateMigration(name, "sql", cfg.MigrationsDir, time.Now())
						if err != nil {
							return err
						}

						fmt.Printf("generate file %s\n", file)
						return nil
					},
				},
				{
					Name:    "locale",
					Usage:   "generate locale file",
					Aliases: []string{"l"},
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name, n",
							Usage: "locale name",
						},
					},
					Action: func(c *cli.Context) error {
						name := c.String("name")
						if len(name) == 0 {
							cli.ShowCommandHelp(c, "locale")
							return nil
						}
						lng, err := language.Parse(name)
						if err != nil {
							return err
						}
						const root = "locales"
						if err = os.MkdirAll(root, 0700); err != nil {
							return err
						}
						file := path.Join(root, fmt.Sprintf("%s.ini", lng.String()))
						fmt.Printf("generate file %s\n", file)
						fd, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
						if err != nil {
							return err
						}
						defer fd.Close()
						return nil
					},
				},
			},
		},
		cli.Command{
			Name:    "routes",
			Aliases: []string{"rt"},
			Usage:   "print out all defined routes",
			Action: func(*cli.Context) error {
				tpl := "%-7s %s\n"
				fmt.Printf(tpl, "METHOD", "PATH")
				for _, r := range Router().Routes() {
					fmt.Printf(tpl, r.Method, r.Path)
				}
				return nil
			},
		},
	)
}

func writeSitemap(root string) error {
	sm := stm.NewSitemap()
	sm.SetDefaultHost(viper.GetString("server.backend"))
	sm.SetPublicPath(root)
	sm.SetCompress(true)
	sm.SetSitemapsPath("/")
	sm.Create()

	for _, u := range _sitemap {
		sm.Add(u)
	}

	if IsProduction() {
		sm.Finalize().PingSearchEngines()
	} else {
		sm.Finalize()
	}
	return nil
}

func writeRssAtom(root string, lang string) error {

	feed := atom.Feed{
		Title:   T(lang, "site.title"),
		ID:      uuid.New().String(),
		Updated: atom.Time(time.Now()),
		Author: &atom.Person{
			Name:  T(lang, "site.author.name"),
			Email: T(lang, "site.author.email"),
		},
		Entry: make([]*atom.Entry, 0),
	}
	home := viper.GetString("server.backend")

	for _, it := range _rss {
		for i := range it.Link {
			it.Link[i].Href = home + it.Link[i].Href
		}
		feed.Entry = append(feed.Entry, it)
	}

	fn := path.Join(root, fmt.Sprintf("rss-%s.atom", lang))
	log.Infof("generate file %s", fn)
	fd, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer fd.Close()
	enc := xml.NewEncoder(fd)
	return enc.Encode(feed)

}

func writeRobotsTxt(root string) error {

	fn := path.Join(root, "robots.txt")
	log.Infof("generate file %s", fn)
	fd, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer fd.Close()
	tpl, err := template.ParseFiles(path.Join("templates", "robots.txt"))
	if err != nil {
		return err
	}
	return tpl.Execute(fd, struct {
		Home string
	}{Home: viper.GetString("server.backend")})
}

func writeGoogleVerify(root string) error {
	var code string
	if err := Get("site.google.verify.code", &code); err != nil {
		return err
	}
	fn := path.Join(root, fmt.Sprintf("google%s.html", code))
	log.Infof("generate file %s", fn)
	fd, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer fd.Close()
	_, err = fmt.Fprintf(fd, "google-site-verification: google%s.html", code)
	return err

}

func writeBaiduVerify(root string) error {
	var code string
	if err := Get("site.baidu.verify.code", &code); err != nil {
		return err
	}
	fn := path.Join(root, fmt.Sprintf("baidu_verify_%s.html", code))
	log.Infof("generate file %s", fn)
	fd, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer fd.Close()
	_, err = fd.WriteString(code)
	return err
}
