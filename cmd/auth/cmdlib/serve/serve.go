package serve

import (
	"database/sql"

	"github.com/andrebq/auth"
	"github.com/andrebq/auth/api"
	"github.com/andrebq/auth/internal/httpserver"
	"github.com/urfave/cli/v2"
)

func Cmd(dir *string) *cli.Command {
	var db *sql.DB
	return &cli.Command{
		Name:  "serve",
		Usage: "Controls the various servers that auth can provide, see sub-commands for more details",
		Subcommands: []*cli.Command{
			serveApiCmd(&db),
		},
		Before: func(ctx *cli.Context) error {
			var err error
			db, err = auth.OpenDir(ctx.Context, *dir)
			return err
		},
		After: func(ctx *cli.Context) error {
			if db != nil {
				db.Close()
			}
			return nil
		},
	}
}

func serveApiCmd(db **sql.DB) *cli.Command {
	port := uint(18001)
	addr := "127.0.0.1"
	return &cli.Command{
		Name:  "api",
		Usage: "Serve the internal API (ie, not exposed to public internet) which is used by other clients to authenticate users",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:        "port",
				Usage:       "Port to bind and listen for incoming connections",
				Destination: &port,
				EnvVars:     []string{"AUTH_SERVE_API_PORT"},
				Value:       port,
			},
			&cli.StringFlag{
				Name:        "bind",
				Usage:       "Address to bind and listen for incoming connections",
				Destination: &addr,
				EnvVars:     []string{"AUTH_SERVE_API_ADDR"},
				Value:       addr,
			},
		},
		Action: func(ctx *cli.Context) error {
			handler := api.Handler(*db)
			return httpserver.Run(ctx.Context, addr, port, handler)
		},
	}
}
