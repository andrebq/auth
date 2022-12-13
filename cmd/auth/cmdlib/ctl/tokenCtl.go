package ctl

import (
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/andrebq/auth"
	"github.com/urfave/cli/v2"
)

func tokenCtlCmd(dir *string, output io.Writer) *cli.Command {
	var db *sql.DB
	return &cli.Command{
		Name:  "token",
		Usage: "Controls tokens for users",
		Subcommands: []*cli.Command{
			registerTokenCmd(&db, output),
			revokeTokenCmd(&db),
		},
		Before: func(ctx *cli.Context) error {
			var err error
			db, err = auth.OpenDir(ctx.Context, *dir)
			if err != nil {
				return err
			}
			return nil
		},
		After: func(ctx *cli.Context) error {
			return db.Close()
		},
	}
}

func registerTokenCmd(db **sql.DB, output io.Writer) *cli.Command {
	var login, tokenType string
	var ttl time.Duration
	var skipnl bool
	return &cli.Command{
		Name:  "register",
		Usage: "Register a new token and prits it to stdout with a newline",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "login",
				Usage:       "User login that is associated with this token",
				EnvVars:     []string{"AUTH_CTL_USERNAME"},
				Destination: &login,
				Required:    true,
			},
			&cli.BoolFlag{
				Name:        "n",
				Usage:       "Disable the newline",
				Destination: &skipnl,
			},
			&cli.DurationFlag{
				Name:        "ttl",
				Usage:       "TTL for the new token",
				Required:    true,
				Destination: &ttl,
			},
			&cli.StringFlag{
				Name:        "token-type",
				Usage:       "Type of token to use",
				Required:    true,
				Destination: &tokenType,
			},
		},
		Action: func(ctx *cli.Context) error {
			token, err := auth.CreateToken(ctx.Context, *db, login, tokenType, time.Now().Add(ttl))
			if err != nil {
				return err
			}
			fmt.Fprint(output, token)
			if !skipnl {
				fmt.Fprintln(output)
			}
			return nil
		},
	}
}

func revokeTokenCmd(db **sql.DB) *cli.Command {
	var tokenID string
	return &cli.Command{
		Name:  "revoke",
		Usage: "Revoke a token by its ID",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "token-id",
				Usage:       "ID of the token to be revoked",
				Destination: &tokenID,
				Required:    true,
			},
		},
		Action: func(ctx *cli.Context) error {
			actualID, err := auth.ExtractTokenID(tokenID)
			if err != nil {
				return err
			}
			return auth.RevokeToken(ctx.Context, *db, actualID)
		},
	}
}
