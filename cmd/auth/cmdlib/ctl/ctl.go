package ctl

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/andrebq/auth"
	"github.com/urfave/cli/v2"
)

func Cmd(dir *string, input io.Reader) *cli.Command {
	return &cli.Command{
		Name:  "ctl",
		Usage: "Controls the auth database",
		Subcommands: []*cli.Command{
			registerUserCmd(dir, input),
		},
	}
}

func registerUserCmd(dir *string, input io.Reader) *cli.Command {
	var login, password string
	return &cli.Command{
		Name:  "register",
		Usage: "Register a new user",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "login",
				Usage:       "User login",
				EnvVars:     []string{"AUTH_CTL_USERNAME"},
				Destination: &login,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "password",
				Usage:       "Password",
				EnvVars:     []string{"AUTH_CTL_PASSWORD"},
				Destination: &password,
				Value:       "-",
				Required:    false,
				Hidden:      true,
			},
		},
		Action: func(ctx *cli.Context) error {
			if password == "-" {
				aux, err := ioutil.ReadAll(input)
				if err != nil {
					return err
				}
				aux = bytes.TrimSpace(aux)
				password = string(aux)
			}
			db, err := auth.OpenDir(ctx.Context, *dir)
			if err != nil {
				return err
			}
			_, err = auth.RegisterUser(ctx.Context, db, login, []byte(password))
			return err
		},
	}
}
