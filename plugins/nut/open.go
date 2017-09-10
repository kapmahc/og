package nut

import (
	"crypto/aes"
	"fmt"
	"log/syslog"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

var _db *gorm.DB
var _redis *redis.Pool

// Redis redis pool
func Redis() *redis.Pool {
	return _redis
}

// DB gorm database
func DB() *gorm.DB {
	return _db
}

// Open open database and redis
func Open(f cli.ActionFunc) cli.ActionFunc {
	return func(c *cli.Context) error {
		// read config file
		if err := viper.ReadInConfig(); err != nil {
			return err
		}

		if IsProduction() {
			// ----------
			log.SetLevel(log.InfoLevel)
			if wrt, err := syslog.New(syslog.LOG_INFO, viper.GetString("server.name")); err == nil {
				log.AddHook(&logrus_syslog.SyslogHook{Writer: wrt})
			} else {
				log.Error(err)
			}
		} else {
			log.SetLevel(log.DebugLevel)
		}

		log.Infof("read config from %s", viper.ConfigFileUsed())
		// open database
		db, err := gorm.Open(viper.GetString("database.driver"), DataSource())
		if err != nil {
			return err
		}
		db.LogMode(true)
		if err = db.DB().Ping(); err != nil {
			return err
		}
		db.DB().SetMaxIdleConns(viper.GetInt("database.pool.max_idle"))
		db.DB().SetMaxOpenConns(viper.GetInt("database.pool.max_open"))
		_db = db
		// init security
		if _aesBlock, err = aes.NewCipher([]byte(viper.GetString("secrets.aes"))); err != nil {
			return err
		}
		_hmacKey = []byte(viper.GetString("secrets.hmac"))
		_jwtKey = []byte(viper.GetString("secrets.jwt"))
		// open redis
		_redis = &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, e := redis.Dial(
					"tcp",
					fmt.Sprintf(
						"%s:%d",
						viper.GetString("redis.host"),
						viper.GetInt("redis.port"),
					),
				)
				if e != nil {
					return nil, e
				}
				if _, e = c.Do("SELECT", viper.GetInt("redis.db")); e != nil {
					c.Close()
					return nil, e
				}
				return c, nil
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		}
		// load locales
		if err := loadLocales("locales"); err != nil {
			return err
		}
		return f(c)
	}
}
