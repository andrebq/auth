package wsutil_test

import (
	"bytes"
	"context"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/andrebq/auth/internal/wsutil"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func TestHub(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	hub := wsutil.NewHub()
	go hub.Run(ctx)
	server := httptest.NewServer(hub)
	defer server.Close()
	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"

	root, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	alice := dial(t, u, uuid.NewSHA1(root, []byte("alice")))
	go alice.Loop()

	bob := dial(t, u, uuid.NewSHA1(root, []byte("bob")))
	go bob.Loop()

	tom := dial(t, u, uuid.NewSHA1(root, []byte("tom")))
	go tom.Loop()

	id := bob.ID()

	alice.Write() <- append(id[:], []byte("hello")...)
	select {
	case <-ctx.Done():
		t.Fatal("I give up!")
	case buf := <-bob.Read():
		if !bytes.HasSuffix(buf, []byte("hello")) {
			t.Fatal("wtf!")
		}
	case <-tom.Read():
		t.Fatal("Nobody sent messages to tom!")
	}
	select {
	case <-ctx.Done():
	case buf := <-bob.Read():
		if !bytes.HasSuffix(buf, []byte("hello")) {
			t.Fatal("wtf!")
		}
	case <-tom.Read():
		t.Fatal("Nobody sent messages to tom!")
	}
}

func dial(t *testing.T, u *url.URL, id uuid.UUID) *wsutil.Chan {
	copy := *u
	qs := copy.Query()
	qs.Add("endpoint_id", id.String())
	copy.RawQuery = qs.Encode()
	conn, _, err := websocket.DefaultDialer.Dial(copy.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	return wsutil.NewChan(conn, id)
}
