package proxy

import (
	"github.com/andrebq/auth/internal/httpserver"
	"github.com/andrebq/auth/proxy"
	"github.com/urfave/cli/v2"
)

func Cmd() *cli.Command {
	var upstream string
	var authEndpoint string = "http://localhost:18001"
	var bind string = "localhost"
	var port uint = 18002
	return &cli.Command{
		Name:  "proxy",
		Usage: "Proxy requets to enforce authentication via cookies",
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:        "port",
				Usage:       "Port to bind for incoming connections",
				Destination: &port,
				Value:       port,
			},
			&cli.StringFlag{
				Name:        "addr",
				Usage:       "Address to bind for incoming connections",
				Destination: &bind,
				Value:       bind,
			},
			&cli.StringFlag{
				Name:        "upstream",
				Usage:       "URL base where requests will be proxied to",
				Destination: &upstream,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "auth-endpoint",
				Usage:       "Endpoint where auth api is listening",
				Destination: &authEndpoint,
				Value:       authEndpoint,
			},
		},
		Action: func(ctx *cli.Context) error {
			handler, err := proxy.Handler(upstream, authEndpoint)
			if err != nil {
				return err
			}
			return httpserver.RunProxy(ctx.Context, bind, port, handler)
		},
	}
}
