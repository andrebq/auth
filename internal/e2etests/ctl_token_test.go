package e2etests

import (
	"bytes"
	"context"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/andrebq/auth"
	"github.com/andrebq/auth/cmd/auth/cmdlib"
)

func TestToken(t *testing.T) {
	ctx := context.Background()
	tmpdir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	db, err := auth.OpenDir(ctx, tmpdir)
	if err != nil {
		t.Fatal(err)
	}
	var uid string
	if uid, err = auth.RegisterUser(ctx, db, "bob", []byte("old-password")); err != nil {
		t.Fatal(err)
	}
	db.Close()
	output := &bytes.Buffer{}
	tokenType := "session"
	args := []string{"auth", "-d", tmpdir, "ctl", "token", "register", "--login", "bob", "--token-type", tokenType, "--ttl", "1s", "-n"}
	app := cmdlib.NewApp(output, bytes.NewBuffer(nil))
	err = app.RunContext(ctx, args)
	if err != nil {
		t.Fatal(err)
	}

	db, err = auth.OpenDir(ctx, tmpdir)
	if err != nil {
		t.Fatal(err)
	}
	if actualUID, actualTokenType, err := auth.TokenLogin(ctx, db, output.String()); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(actualUID, uid) {
		t.Fatalf("Invalid UID, should be %v got %v", uid, actualUID)
	} else if !reflect.DeepEqual(actualTokenType, tokenType) {
		t.Fatalf("Invalid token type, should be %v got %v", tokenType, actualTokenType)
	}
	db.Close()

	tokenUID, err := auth.ExtractTokenID(output.String())
	if err != nil {
		t.Fatal(err)
	}

	args = []string{"auth", "-d", tmpdir, "ctl", "token", "revoke", "--token-id", tokenUID}
	app = cmdlib.NewApp(io.Discard, bytes.NewBuffer(nil))
	err = app.RunContext(ctx, args)
	if err != nil {
		t.Fatal(err)
	}

	db, err = auth.OpenDir(ctx, tmpdir)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := auth.TokenLogin(ctx, db, output.String()); err == nil {
		t.Fatal("Token was revoked and login using it should fail")
	}
	db.Close()
}
