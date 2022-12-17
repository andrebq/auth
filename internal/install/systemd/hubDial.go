package systemd

import (
	"io"
	"path"
	"path/filepath"
	"text/template"

	"github.com/urfave/cli/v2"
)

type (
	HubDialConfig struct {
		Binary      string
		Description string
		HubEndpoint string
		Token       string
		TunnelID    string
		Bind        string
		Port        uint
	}
)

var (
	hubDial = template.Must(template.New("root").Parse(`
{{ define "unit" }}
[Unit]
Description={{.Description}}

[Service]
ExecStart={{.Binary}} hub dial --hub {{.HubEndpoint}} --bind {{.Bind}} --port {{.Port}} --token {{.Token}} -t {{.TunnelID}}
Restart=always

[Install]
WantedBy=multi-user.target
{{ end }}
	`))
)

func (h *HubDialConfig) Render(out io.Writer) error {
	h.setDefaults()
	return renderTemplate(out, hubDial, "unit", h)
}

func (h *HubDialConfig) setDefaults() {
	if h.HubEndpoint == "" {
		h.HubEndpoint = "ws://localhost:18001"
	}
	if h.Binary == "" {
		h.Binary = filepath.FromSlash(path.Join("/", "usr", "local", "bin", "auth"))
	}
}

func (h *HubDialConfig) Flags() []cli.Flag {
	h.setDefaults()
	return []cli.Flag{
		stringFlag(&h.Description, "description", "Description of the unit service"),
		stringFlag(&h.Binary, "binary", "Path to auth binary"),
		stringFlag(&h.HubEndpoint, "hub-endpoint", "Address where hub is running"),
		stringFlag(&h.Token, "token", "Token to use when authenticating to auth"),
		stringFlag(&h.TunnelID, "tunnel", "ID of the tunnel to establish"),
		stringFlag(&h.Bind, "bind", "Address where the dialer will listen for incoming connections"),
		uintFlag(&h.Port, "port", "Port where the dialer will listen for incoming connections"),
	}
}
