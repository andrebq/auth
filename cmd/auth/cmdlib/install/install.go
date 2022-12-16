package install

import (
	"io"

	"github.com/andrebq/auth/internal/install/systemd"
	"github.com/urfave/cli/v2"
)

func Cmd(output io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "Generate installation scripts for various auth components",
		Subcommands: []*cli.Command{
			systemdCmd(output),
		},
	}
}

func systemdCmd(output io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "systemd",
		Usage: "Generate systemd unit files and prints them to stdout",
		Subcommands: []*cli.Command{
			authCmd(output),
		},
	}
}

func authCmd(output io.Writer) *cli.Command {
	ac := systemd.AuthConfig{}
	return &cli.Command{
		Name:  "auth",
		Usage: "Generate a unit service for the auth command",
		Flags: ac.Flags(),
		Action: func(ctx *cli.Context) error {
			return ac.Render(output)
		},
	}
}
