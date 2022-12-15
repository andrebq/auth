package cmdlib

import (
	"io"
	"path"
	"path/filepath"

	"github.com/andrebq/auth/cmd/auth/cmdlib/ctl"
	"github.com/andrebq/auth/cmd/auth/cmdlib/hub"
	"github.com/andrebq/auth/cmd/auth/cmdlib/proxy"
	"github.com/andrebq/auth/cmd/auth/cmdlib/serve"
	"github.com/urfave/cli/v2"
)

func NewApp(output io.Writer, input io.Reader) *cli.App {
	var dir string
	return &cli.App{
		Name:  "auth",
		Usage: "Runs/controls the auth server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "data-dir",
				Aliases:     []string{"d"},
				EnvVars:     []string{"AUTH_DATA_DIR"},
				Destination: &dir,
				Value:       filepath.FromSlash(path.Join("/", "var", "authdb", "data-dir")),
			},
		},
		Commands: []*cli.Command{
			ctl.Cmd(&dir, output, input),
			serve.Cmd(&dir),
			proxy.Cmd(),
			hub.Cmd(),
		},
	}
}
