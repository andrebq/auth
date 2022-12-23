package systemd

import (
	"io"
	"path"
	"path/filepath"
	"text/template"

	"github.com/urfave/cli/v2"
)

type (
	ProxyConfig struct {
		Description  string
		Binary       string
		Port         uint
		Bind         string
		AuthEndpoint string
		Upstream     string
	}
)

var (
	proxyUnit = template.Must(template.New("root").Parse(`
{{ define "unit" }}
[Unit]
Description={{.Description}}

[Service]
ExecStart={{.Binary}} proxy --upstream {{.Upstream}} --port {{.Port}} --addr {{.Bind}} --auth-endpoint {{.AuthEndpoint}}
Restart=always

[Install]
WantedBy=multi-user.target
{{ end }}
	`))
)

func (p *ProxyConfig) Render(out io.Writer) error {
	p.setDefaults()
	return renderTemplate(out, proxyUnit, "unit", p)
}

func (p *ProxyConfig) Flags() []cli.Flag {
	p.setDefaults()
	return []cli.Flag{
		stringFlag(&p.Description, "description", "Description of this proxy"),
		stringFlag(&p.AuthEndpoint, "auth-endpoint", "Endpoint where auth API is running"),
		stringFlag(&p.Binary, "binary", "Path to auth binary"),
		stringFlag(&p.Upstream, "upstream", "Upstream server to proxy requests to"),
		stringFlag(&p.Bind, "bind", "Address where the proxy will listen for requests"),
		uintFlag(&p.Port, "port", "Port where the proxy will listen for requests"),
	}
}

func (p *ProxyConfig) setDefaults() {
	if p.AuthEndpoint == "" {
		p.AuthEndpoint = "http://localhost:18001/"
	}
	if p.Binary == "" {
		p.Binary = filepath.FromSlash(path.Join("/", "usr", "local", "bin", "auth"))
	}
	if p.Bind == "" {
		p.Bind = "localhost"
	}
	if p.Port == 0 {
		p.Port = 18002
	}
}
