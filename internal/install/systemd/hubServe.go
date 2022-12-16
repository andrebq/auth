package systemd

import (
	"io"
	"path"
	"path/filepath"
	"text/template"

	"github.com/urfave/cli/v2"
)

type (
	HubServeConfig struct {
		Description  string
		Binary       string
		AuthEndpoint string
		Bind         string
		Port         uint
	}
)

var (
	hubServe = template.Must(template.New("root").Parse(`
{{ define "unit" }}
[Unit]
Description={{.Description}}

[Service]
ExecStart={{.Binary}} hub serve --auth-endpoint {{.AuthEndpoint}} --bind {{.Bind}} --port {{.Port}}
Restart=true

[Install]
WantedBy=multi-user.target
{{ end }}
	`))
)

func (h *HubServeConfig) Render(out io.Writer) error {
	h.setDefaults()
	return renderTemplate(out, hubServe, "unit", h)
}

func (h *HubServeConfig) setDefaults() {
	if h.Description == "" {
		h.Description = "Runs the auth hub"
	}
	if h.AuthEndpoint == "" {
		h.AuthEndpoint = "http://localhost:18001"
	}
	if h.Bind == "" {
		h.Bind = "localhost"
	}
	if h.Port == 0 {
		h.Port = 18003
	}
	if h.Binary == "" {
		h.Binary = filepath.FromSlash(path.Join("/", "usr", "local", "bin", "auth"))
	}
}

func (h *HubServeConfig) Flags() []cli.Flag {
	h.setDefaults()
	return []cli.Flag{
		stringFlag(&h.Binary, "binary", "Path to auth binary"),
		stringFlag(&h.Description, "description", "Description of the unit service"),
		stringFlag(&h.AuthEndpoint, "auth-endpoint", "Endpoint where auth is running"),
		stringFlag(&h.Bind, "bind", "Address to listen for incoming requests"),
		uintFlag(&h.Port, "port", "Port to listen for incoming requests"),
	}
}
