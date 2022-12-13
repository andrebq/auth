package ctl

import (
	"bytes"
	"io"

	"github.com/andrebq/auth"
	"github.com/urfave/cli/v2"
)

func Cmd(dir *string, output io.Writer, input io.Reader) *cli.Command {
	return &cli.Command{
		Name:  "ctl",
		Usage: "Controls the auth database",
		Subcommands: []*cli.Command{
			registerUserCmd(dir, input),
			updateUserCmd(dir, input),
			tokenCtlCmd(dir, output),
		},
	}
}

func updateUserCmd(dir *string, input io.Reader) *cli.Command {
	var login, password string
	return &cli.Command{
		Name:  "passwd",
		Usage: "Update the user password",
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
				aux, err := io.ReadAll(input)
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
			return auth.ReplacePassword(ctx.Context, db, login, []byte(password))
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
				aux, err := io.ReadAll(input)
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
