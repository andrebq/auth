package wsutil_test

import (
	"bytes"
	"context"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/andrebq/auth/internal/wsutil"
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

	alice := dial(t, u.String())
	go alice.Loop()

	bob := dial(t, u.String())
	go bob.Loop()

	alice.Write() <- []byte("hello")
	select {
	case <-ctx.Done():
		t.Fatal("I give up!")
	case buf := <-bob.Read():
		if !bytes.Equal(buf, []byte("hello")) {
			t.Fatal("wtf!")
		}
	}
}

func dial(t *testing.T, str string) *wsutil.Chan {
	conn, _, err := websocket.DefaultDialer.Dial(str, nil)
	if err != nil {
		t.Fatal(err)
	}
	return wsutil.NewChan(conn)
}
