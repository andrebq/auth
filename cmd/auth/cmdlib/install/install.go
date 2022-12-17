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
			unitFileCmd(output),
			authCmd(output),
			proxyCmd(output),
			hubCmd(output),
		},
	}
}

func unitFileCmd(output io.Writer) *cli.Command {
	ufc := systemd.UnitFileConfig{}
	return &cli.Command{
		Name:  "unit-file",
		Usage: "Render the path to a unit file with the given name",
		Flags: ufc.Flags(),
		Action: func(ctx *cli.Context) error {
			err := ufc.Render(output)
			io.WriteString(output, "\n")
			return err
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

func proxyCmd(output io.Writer) *cli.Command {
	pc := systemd.ProxyConfig{}
	return &cli.Command{
		Name:  "proxy",
		Usage: "Generate a unit service for the auth proxy command",
		Flags: pc.Flags(),
		Action: func(ctx *cli.Context) error {
			return pc.Render(output)
		},
	}
}

func hubCmd(output io.Writer) *cli.Command {
	return &cli.Command{
		Name:  "hub",
		Usage: "Generate unit-services for hub subcommands",
		Subcommands: []*cli.Command{
			hubExposeCmd(output),
			hubDialCmd(output),
			hubServeCmd(output),
		},
	}
}

func hubServeCmd(output io.Writer) *cli.Command {
	hsc := systemd.HubServeConfig{}
	return &cli.Command{
		Name:  "serve",
		Usage: "Generate unit-services for hub serve",
		Flags: hsc.Flags(),
		Action: func(ctx *cli.Context) error {
			return hsc.Render(output)
		},
	}
}

func hubExposeCmd(output io.Writer) *cli.Command {
	hec := systemd.HubExposeConfig{}
	return &cli.Command{
		Name:  "expose",
		Usage: "Generate unit-services for hub expose",
		Flags: hec.Flags(),
		Action: func(ctx *cli.Context) error {
			return hec.Render(output)
		},
	}
}

func hubDialCmd(output io.Writer) *cli.Command {
	hdc := systemd.HubDialConfig{}
	return &cli.Command{
		Name:  "dial",
		Usage: "Generate unit-services for hub dial",
		Flags: hdc.Flags(),
		Action: func(ctx *cli.Context) error {
			return hdc.Render(output)
		},
	}
}
