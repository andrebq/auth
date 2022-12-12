package auth_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/andrebq/auth"
)

func TestTokenAuth(t *testing.T) {
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
	tokenType := "session"
	passwd := []byte("super-secure")

	uid, err := auth.RegisterUser(ctx, db, login, passwd)
	if err != nil {
		t.Fatal(err)
	}

	token, err := auth.CreateToken(ctx, db, login, tokenType, time.Now().Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	actualUID, actualTokenType, err := auth.TokenLogin(ctx, db, token)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actualUID, uid) {
		t.Errorf("UID should be %v got %v", uid, actualUID)
	}
	if !reflect.DeepEqual(actualTokenType, tokenType) {
		t.Errorf("Token Type should be %v got %v", tokenType, actualTokenType)
	}

	tid, err := auth.ExtractTokenID(token)
	if err != nil {
		t.Fatal(err)
	}

	err = auth.RevokeToken(ctx, db, tid)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = auth.TokenLogin(ctx, db, token)
	if err == nil {
		t.Fatal("Token has been revoked but can still login!")
	}
}
