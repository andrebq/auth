package hub

import (
	"context"
	"fmt"
	"net"

	"github.com/andrebq/auth/internal/httpserver"
	"github.com/andrebq/auth/tunnel/hub"
	"github.com/andrebq/auth/tunnel/hub/proxy"
	"github.com/urfave/cli/v2"
)

func Cmd() *cli.Command {
	return &cli.Command{
		Name:  "hub",
		Usage: "Contains commands to operate a tunnel hub",
		Subcommands: []*cli.Command{
			serveCmd(),
			exposeLocalCmd(),
			dialRemoteCmd(),
		},
	}
}

func serveCmd() *cli.Command {
	var authEndpoint string = "http://localhost:18001/"
	var addr string = "127.0.0.1"
	var port uint = 18003
	return &cli.Command{
		Name:  "serve",
		Usage: "Runs the hub server that creates tunnels",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "auth-endpoint",
				Aliases:     []string{"ae"},
				Usage:       "Endpoint where the auth server is running",
				Destination: &authEndpoint,
				Value:       authEndpoint,
			},
			&cli.StringFlag{
				Name:        "bind",
				Usage:       "Address where the hub will listen",
				Destination: &addr,
				Value:       addr,
			},
			&cli.UintFlag{
				Name:        "port",
				Usage:       "Port where the hub will listen",
				Destination: &port,
				Value:       port,
			},
		},
		Action: func(ctx *cli.Context) error {
			h, err := hub.NewHub()
			if err != nil {
				return err
			}
			return httpserver.RunProxy(ctx.Context, addr, port, h)
		},
	}
}

func exposeLocalCmd() *cli.Command {
	var localAddr string
	var token string
	var tunnelID string
	hub := "ws://localhost:18003/"
	return &cli.Command{
		Name:  "expose",
		Usage: "Expose a local port on a remote hub under a given tunnel ID",
		Description: `When started, this process will dial to a given tunnel hub
and Accept clients.

Each new tunnel client will then be proxied to the local server`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "local-addr",
				Usage:       "Local Address where to dial",
				Destination: &localAddr,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "token",
				Usage:       "Token to authenticate",
				EnvVars:     []string{"AUTH_HUB_TOKEN"},
				Hidden:      true,
				Required:    true,
				Destination: &token,
			},
			&cli.StringFlag{
				Name:        "tunnel-id",
				Usage:       "ID of the tunnel",
				Aliases:     []string{"t"},
				Required:    true,
				Destination: &tunnelID,
			},
			&cli.StringFlag{
				Name:        "hub",
				Usage:       "Hub address",
				Value:       hub,
				Destination: &hub,
			},
		},
		Action: func(ctx *cli.Context) error {
			dialer := func(_ context.Context) (net.Conn, error) {
				return net.Dial("tcp", localAddr)
			}
			return proxy.RemoteToLocal(ctx.Context, hub, token, tunnelID, dialer)
		},
	}
}

func dialRemoteCmd() *cli.Command {
	var token, tunnelID string
	hub := "ws://localhost:18003/"
	var addr string
	var port uint
	return &cli.Command{
		Name:  "dial",
		Usage: "Accept connections on a given local listener and dials to a given tunnel",
		Description: `When started, this process will start a local-listener, for each new
connection made, it will then dial to a tunnel on the given hub.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "bind",
				Usage:       "Local address to listen for incoming connections",
				Destination: &addr,
				Required:    true,
			},
			&cli.UintFlag{
				Name:        "port",
				Usage:       "Port to listen for incoming connections",
				Destination: &port,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "token",
				Usage:       "Token to authenticate",
				EnvVars:     []string{"AUTH_HUB_TOKEN"},
				Hidden:      true,
				Required:    true,
				Destination: &token,
			},
			&cli.StringFlag{
				Name:        "tunnel-id",
				Usage:       "ID of the tunnel",
				Aliases:     []string{"t"},
				Required:    true,
				Destination: &tunnelID,
			},
			&cli.StringFlag{
				Name:        "hub",
				Usage:       "Hub address",
				Value:       hub,
				Destination: &hub,
			},
		},
		Action: func(ctx *cli.Context) error {
			lst, err := net.Listen("tcp", fmt.Sprintf("%v:%v", addr, port))
			if err != nil {
				return err
			}
			return proxy.LocalToRemote(ctx.Context, lst, hub, token, tunnelID)
		},
	}
}
