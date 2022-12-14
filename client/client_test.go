package client_test

import (
	"context"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/andrebq/auth"
	"github.com/andrebq/auth/api"
	"github.com/andrebq/auth/client"
)

func TestClientAPI(t *testing.T) {
	ctx := context.Background()
	db, err := auth.OpenMemory(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	server := httptest.NewServer(api.Handler(db))
	defer server.Close()

	var uid string
	if uid, err = auth.RegisterUser(ctx, db, "bob", []byte("bob")); err != nil {
		t.Fatal(err)
	}

	cli := client.New(server.URL)
	if err := cli.Login(ctx, "bob", "bob"); err != nil {
		t.Fatal(err)
	}

	token, err := cli.StartSession(ctx, "bob", "bob", time.Second)
	if err != nil {
		t.Fatal(err)
	} else if _, err = auth.ExtractTokenID(token); err != nil {
		t.Fatal(err)
	}

	actualUID, actualTokenType, err := cli.ValidateToken(ctx, token)
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(actualUID, uid) {
		t.Fatalf("Invalid uid, should be %v got %v", uid, actualUID)
	} else if !reflect.DeepEqual(actualTokenType, "session") {
		t.Fatalf("Invalid token type should be session got %v", actualTokenType)
	}
}
