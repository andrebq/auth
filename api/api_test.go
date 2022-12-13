package api_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/andrebq/auth"
	"github.com/andrebq/auth/api"
	"github.com/steinfletcher/apitest"
	jsonpath "github.com/steinfletcher/apitest-jsonpath"
)

func TestLoginProcess(t *testing.T) {
	ctx := context.Background()
	db, err := auth.OpenMemory(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := auth.RegisterUser(ctx, db, "bob", []byte("1234")); err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	apitest.Handler(api.Handler(db)).
		Post("/auth/login").
		Body(`{"login":"bob", "password": "1234"}`).
		Expect(t).
		Status(http.StatusOK).
		End()
}

func TestTokenAuth(t *testing.T) {
	ctx := context.Background()
	db, err := auth.OpenMemory(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var uid string
	if uid, err = auth.RegisterUser(ctx, db, "bob", []byte("1234")); err != nil {
		t.Fatal(err)
	}
	var token string
	var tokenID string
	if token, err = auth.CreateToken(ctx, db, "bob", "session", time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	tokenID, _ = auth.ExtractTokenID(token)
	defer db.Close()
	apitest.Handler(api.Handler(db)).
		Post("/auth/token").
		Bodyf(`{"token":%q}`, token).
		Expect(t).
		Status(http.StatusOK).
		Assert(jsonpath.Chain().Equal("tokenID", tokenID).Equal("uid", uid).Equal("tokenType", "session").End()).
		End()
}

func TestSession(t *testing.T) {
	ctx := context.Background()
	db, err := auth.OpenMemory(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var uid string
	if uid, err = auth.RegisterUser(ctx, db, "bob", []byte("1234")); err != nil {
		t.Fatal(err)
	}
	var session struct {
		Token string `json:"token"`
	}
	apitest.Handler(api.Handler(db)).
		Post("/session").
		Body(`{"login":"bob", "password": "1234", "ttl": "1s"}`).
		Expect(t).
		Status(http.StatusOK).
		Assert(jsonpath.Present("token")).
		End().JSON(&session)

	actualUID, actualTokenType, err := auth.TokenLogin(ctx, db, session.Token)
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(actualUID, uid) {
		t.Fatalf("Invalid UID, should be %v got %v", uid, actualUID)
	} else if !reflect.DeepEqual("session", actualTokenType) {
		t.Fatalf("Token type should be session but got %v", actualTokenType)
	}
	time.Sleep(time.Second)
	_, _, err = auth.TokenLogin(ctx, db, session.Token)
	if err == nil {
		t.Fatalf("Token should have a ttl of just 100ms, but it is still valid after 100ms")
	}
}
