package systemd

import (
	"errors"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"regexp"

	"github.com/urfave/cli/v2"
)

type (
	UnitFileConfig struct {
		Name      string
		SystemDir string
	}
)

var (
	validUnitName = regexp.MustCompile(`^[\w-\.]+$`)
)

func (u *UnitFileConfig) Flags() []cli.Flag {
	u.setDefaults()
	return []cli.Flag{
		stringFlag(&u.Name, "name", "Name of unit file (ie.: systemctl start <name>)"),
		stringFlag(&u.SystemDir, "system-dir", "Path to directory where service files should be stored"),
	}
}

func (u *UnitFileConfig) Render(out io.Writer) error {
	u.setDefaults()
	if u.Name == "" {
		return errors.New("missing name of unit file")
	}
	if !validUnitName.MatchString(u.Name) {
		return fmt.Errorf("unit-name must comply with the following regex: %v", validUnitName.String())
	}
	_, err := io.WriteString(out, filepath.Join(u.SystemDir, fmt.Sprintf("%v.%v", filepath.Clean(u.Name), "service")))
	return err
}

func (u *UnitFileConfig) setDefaults() {
	if u.SystemDir == "" {
		u.SystemDir = filepath.FromSlash(path.Join("/", "lib", "systemd", "system"))
	}
}
