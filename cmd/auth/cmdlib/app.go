package cmdlib

import (
	"io"
	"path/filepath"

	"github.com/andrebq/auth/cmd/auth/cmdlib/ctl"
	"github.com/urfave/cli/v2"
)

func NewApp(input io.Reader) *cli.App {
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
				Value:       filepath.Join("var", "authdb", "data-dir"),
			},
		},
		Commands: []*cli.Command{
			ctl.Cmd(&dir, input),
		},
	}
}
