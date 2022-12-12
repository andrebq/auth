package e2etests

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/andrebq/auth"
	"github.com/andrebq/auth/cmd/auth/cmdlib"
)

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	db, err := auth.OpenDir(ctx, tmpdir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := auth.RegisterUser(ctx, db, "bob", []byte("old-password")); err != nil {
		t.Fatal(err)
	}
	db.Close()
	input := strings.NewReader("new-password\n")
	app := cmdlib.NewApp(input)
	args := []string{"auth", "-d", tmpdir, "ctl", "passwd", "--login", "bob"}
	err = app.RunContext(ctx, args)
	if err != nil {
		t.Fatal(err)
	}

	db, err = auth.OpenDir(ctx, tmpdir)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err = auth.Login(ctx, db, "bob", []byte("new-password")); err != nil {
		t.Fatal(err)
	}
}
