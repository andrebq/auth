package hub

import (
	"errors"

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
	}
}

func exposeLocalCmd() *cli.Command {
	return &cli.Command{
		Name:  "expose",
		Usage: "Expose a local port on a remote hub under a given tunnel ID",
		Description: `When started, this process will dial to a given tunnel hub
and Accept clients.

Each new tunnel client will then be proxied to the local server`,
		Action: func(ctx *cli.Context) error { return errors.New("not implemented") },
	}
}

func dialRemoteCmd() *cli.Command {
	return &cli.Command{
		Name:  "dial",
		Usage: "Accept connections on a given local listener and dials to a given tunnel",
		Description: `When started, this process will start a local-listener, for each new
connection made, it will then dial to a tunnel on the given hub.`,
		Action: func(ctx *cli.Context) error { return errors.New("not implemented") },
	}
}
