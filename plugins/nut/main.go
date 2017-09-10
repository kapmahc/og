package nut

import (
	"fmt"
	"time"

	"github.com/urfave/cli"
)

var (
	// Version version
	Version string
	// BuildTime build time
	BuildTime string
	// Usage usage
	Usage string
	// Copyright copyright
	Copyright string
	// AuthorName author's name
	AuthorName string
	// AuthorEmail author's email
	AuthorEmail string
)

var _commands []cli.Command

// AddConsoleTask register console tasks
func AddConsoleTask(args ...cli.Command) {
	_commands = append(_commands, args...)
}

// Main entry
func Main(args ...string) error {
	ap := cli.NewApp()
	ap.Name = args[0]
	ap.Version = fmt.Sprintf("%s (%s)", Version, BuildTime)
	ap.Authors = []cli.Author{
		cli.Author{
			Name:  AuthorName,
			Email: AuthorEmail,
		},
	}
	if BuildTime != "" {
		ts, err := time.Parse(time.RFC1123Z, BuildTime)
		if err != nil {
			return err
		}
		ap.Compiled = ts
	}

	ap.Copyright = Copyright
	ap.Usage = Usage
	ap.EnableBashCompletion = true
	ap.Commands = _commands

	return ap.Run(args)
}
