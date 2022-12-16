package systemd

import (
	"io"
	"path"
	"path/filepath"
	"text/template"

	"github.com/urfave/cli/v2"
)

type (
	HubExposeConfig struct {
		Description string
		Binary      string
		HubEndpoint string
		Token       string
		TunnelID    string
		Local       string
	}
)

var (
	hubExpose = template.Must(template.New("root").Parse(`
{{ define "unit" }}
[Unit]
Description={{.Description}}

[Service]
ExecStart={{.Binary}} hub expose --hub {{.HubEndpoint}} --token "{{.Token}}" -t {{.TunnelID}} --local-addr {{.Local}}
Restart=true

[Install]
WantedBy=multi-user.target
{{ end }}
	`))
)

func (h *HubExposeConfig) Render(out io.Writer) error {
	h.setDefaults()
	return renderTemplate(out, hubExpose, "unit", h)
}

func (h *HubExposeConfig) Flags() []cli.Flag {
	h.setDefaults()
	return []cli.Flag{
		stringFlag(&h.Description, "description", "Description of the unit service"),
		stringFlag(&h.Binary, "binary", "Path to auth binary"),
		stringFlag(&h.HubEndpoint, "hub-endpoint", "Endpoint where hub is running"),
		stringFlag(&h.Local, "local", "Local-network address which we want to expose"),
		stringFlag(&h.Token, "token", "Token to use when authenticating to auth"),
		stringFlag(&h.TunnelID, "tunnel", "ID of the tunnel to establish"),
	}
}

func (h *HubExposeConfig) setDefaults() {
	if h.HubEndpoint == "" {
		h.HubEndpoint = "ws://localhost:18001"
	}
	if h.Binary == "" {
		h.Binary = filepath.FromSlash(path.Join("/", "usr", "local", "bin", "auth"))
	}
}
