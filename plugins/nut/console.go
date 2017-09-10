package nut

import (
	"fmt"

	"github.com/urfave/cli"
)

func init() {
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
						if err := _db.Select([]string{"name", "email", "uid"}).
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
	)
}
