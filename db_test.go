package auth_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/andrebq/auth"
)

func TestLogin(t *testing.T) {
	ctx := context.Background()
	db, err := auth.OpenMemory(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = auth.InitDB(ctx, db); err != nil {
		t.Fatal(err)
	}
	login := "bob"
	passwd := []byte("super-secure")

	uid, err := auth.RegisterUser(ctx, db, login, passwd)
	if err != nil {
		t.Fatal(err)
	}
	authUID, err := auth.Login(ctx, db, login, passwd)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(uid, authUID) {
		t.Fatalf("udi and authUID do not match: %v != %v", uid, authUID)
	}
	passwd = []byte("extra-secure")
	if err := auth.ReplacePassword(ctx, db, login, passwd); err != nil {
		t.Fatal(err)
	}
	authUID, err = auth.Login(ctx, db, login, passwd)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(uid, authUID) {
		t.Fatalf("udi and authUID do not match: %v != %v", uid, authUID)
	}
}
