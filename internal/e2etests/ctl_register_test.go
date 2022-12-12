package e2etests

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/andrebq/auth"
	"github.com/andrebq/auth/cmd/auth/cmdlib"
)

func TestRegister(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	input := strings.NewReader("secure-password\n")
	app := cmdlib.NewApp(input)
	args := []string{"auth", "-d", tmpdir, "ctl", "register", "--login", "bob"}
	ctx := context.Background()
	err = app.RunContext(ctx, args)
	if err != nil {
		t.Fatal(err)
	}

	db, err := auth.OpenDir(ctx, tmpdir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err = auth.Login(ctx, db, "bob", []byte("secure-password")); err != nil {
		t.Fatal(err)
	}
}
