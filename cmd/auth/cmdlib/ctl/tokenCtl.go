package ctl

import "github.com/urfave/cli/v2"

func tokenCtlCmd(dir *string) *cli.Command {
	return &cli.Command{
		Name:        "token",
		Usage:       "Controls tokens for users",
		Subcommands: []*cli.Command{},
	}
}
