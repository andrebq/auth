package systemd

import (
	"bytes"
	"io"
	"path"
	"path/filepath"
	"text/template"
	"unicode"

	"github.com/urfave/cli/v2"
)

type (
	AuthConfig struct {
		Datadir     string
		Binary      string
		Port        uint
		Bind        string
		Description string
	}
)

var (
	authUnit = template.Must(template.New("root").Parse(`
{{ define "unit" }}
[Unit]
Description={{.Description}}

[Service]
ExecStart={{.Binary}} --data-dir "{{.Datadir}}" serve api --port {{.Port}} --bind {{.Bind}}
Restart=always

[Install]
WantedBy=multi-user.target
{{ end }}
	`))
)

func (a *AuthConfig) Render(out io.Writer) error {
	a.setDefaults()
	return renderTemplate(out, authUnit, "unit", a)
}

func (a *AuthConfig) Flags() []cli.Flag {
	a.setDefaults()
	return []cli.Flag{
		stringFlag(&a.Datadir, "data-dir", "Where to save logins and tokens"),
		stringFlag(&a.Binary, "binary", "Where the binary is located"),
		stringFlag(&a.Description, "description", "Description of the unit service"),
		stringFlag(&a.Bind, "bind", "Address where to bind"),
		uintFlag(&a.Port, "port", "Port where to bind"),
	}
}

func (a *AuthConfig) setDefaults() {
	if a.Datadir == "" {
		a.Datadir = filepath.FromSlash(path.Join("/", "var", "authdb", "data-dir"))
	}
	if a.Binary == "" {
		a.Binary = filepath.FromSlash(path.Join("/", "usr", "local", "bin", "auth"))
	}
	if a.Description == "" {
		a.Description = "HTTP Auth API"
	}
	if a.Port == 0 {
		a.Port = 18001
	}
	if a.Bind == "" {
		a.Bind = "localhost"
	}
}

func stringFlag(out *string, name string, usage string) *cli.StringFlag {
	return &cli.StringFlag{
		Name:        name,
		Usage:       usage,
		Destination: out,
		Required:    len(*out) == 0,
		Value:       *out,
	}
}

func uintFlag(out *uint, name string, usage string) *cli.UintFlag {
	return &cli.UintFlag{
		Name:        name,
		Usage:       usage,
		Destination: out,
		Required:    (*out == 0),
		Value:       *out,
	}
}

func renderTemplate(out io.Writer, tmpl *template.Template, name string, data interface{}) error {
	buf := bytes.Buffer{}
	err := tmpl.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return err
	}
	_, err = out.Write(bytes.TrimLeftFunc(buf.Bytes(), unicode.IsSpace))
	return err
}
