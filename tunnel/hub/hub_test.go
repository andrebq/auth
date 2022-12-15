package hub_test

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/andrebq/auth/tunnel/hub"
)

func TestHub(t *testing.T) {
	h, err := hub.NewHub()
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(h)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	wsBase := strings.Replace(server.URL, "http://", "ws://", 1)

	gotPing, gotPong := make(chan error, 1), make(chan error, 1)

	tunnelServer := func(out chan error) {
		defer close(out)
		conn, err := hub.Accept(wsBase, "change-for-a-valid-token", "tunnel-01")
		if err != nil {
			out <- err
			return
		}
		conn.SetDeadline(time.Now().Add(time.Minute))
		defer conn.Close()
		msg := make([]byte, 4)
		_, err = io.ReadFull(conn, msg)
		if err != nil {
			out <- err
			return
		}
		_, err = conn.Write(msg)
		if err != nil {
			out <- err
			return
		}
	}

	tunnelClient := func(out chan error) {
		defer close(out)
		conn, err := hub.Dial(wsBase, "change-for-another-valid-token", "tunnel-01")
		if err != nil {
			out <- err
			return
		}
		conn.SetDeadline(time.Now().Add(time.Minute))
		defer conn.Close()
		msg := []byte("ping")
		_, err = conn.Write(msg)
		if err != nil {
			out <- err
			return
		}
		_, err = io.ReadFull(conn, msg)
		if err != nil {
			out <- err
			return
		}
	}

	go tunnelServer(gotPing)
	go tunnelClient(gotPong)

	tick := time.After(time.Second * 2)

	for {
		select {
		case err := <-gotPing:
			if err != nil {
				t.Fatalf("PING failed with: %v", err)
			}
			gotPing = nil
		case err := <-gotPong:
			if err != nil {
				t.Fatalf("PONG failed with: %v", err)
			}
			gotPong = nil
		case <-tick:
			t.Fatal("Timeout happened and system did not finish")
		}

		if gotPing == nil && gotPong == nil {
			// nothing to wait anymore
			break
		}
	}
}
