package hub

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type (
	H struct {
		lock  sync.Mutex
		dials map[string]chan dialState
	}

	dialState struct {
		client *websocket.Conn
		done   chan struct{}
	}
)

var (
	upgrader = websocket.Upgrader{}
)

func NewHub() (http.Handler, error) {
	hub := &H{
		dials: make(map[string]chan dialState),
	}
	mux := http.NewServeMux()
	mux.Handle("/ws/listen", http.HandlerFunc(hub.handleListen))
	mux.Handle("/ws/dial", http.HandlerFunc(hub.handleDial))
	return mux, nil
}

func (h *H) handleListen(w http.ResponseWriter, req *http.Request) {
	tunnelID := req.FormValue("tunnel_id")
	err := h.validToken(getToken(req), tunnelID)
	if err != nil {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, "Upgrade failed", http.StatusBadRequest)
		return
	}
	ctx := req.Context()
	ticker := time.NewTicker(time.Second)
	dials := h.acquireDialer(tunnelID)
	for {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-ticker.C:
			conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Minute))
		case dial := <-dials:
			proxyConn(req.Context(), dial.client, conn)
			close(dial.done)
			return
		}
	}
}

func proxyConn(ctx context.Context, client, server *websocket.Conn) {
	ctx, cancel := context.WithCancel(ctx)
	atob := func(done func(), a, b *websocket.Conn) {
		defer done()
		for {
			a.SetReadDeadline(time.Now().Add(time.Minute))
			mt, buf, err := a.ReadMessage()
			if err != nil {
				return
			}
			b.SetWriteDeadline(time.Now().Add(time.Minute))
			err = b.WriteMessage(mt, buf)
			if err != nil {
				return
			}
		}
	}
	client.WriteMessage(websocket.BinaryMessage, signalPacket)
	server.WriteMessage(websocket.BinaryMessage, signalPacket)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go atob(wg.Done, client, server)
	go atob(wg.Done, server, client)
	go func() {
		<-ctx.Done()
		client.Close()
		server.Close()
	}()
	wg.Wait()
	cancel()
}

func (h *H) acquireDialer(tunnelID string) chan dialState {
	h.lock.Lock()
	defer h.lock.Unlock()
	ch := h.dials[tunnelID]
	if ch == nil {
		ch = make(chan dialState)
		h.dials[tunnelID] = ch
	}
	return ch
}

func (h *H) handleDial(w http.ResponseWriter, req *http.Request) {
	tunnelID := req.FormValue("tunnel_id")
	err := h.validToken(getToken(req), tunnelID)
	if err != nil {
		http.Error(w, "Not authorized", http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, "Upgrade failed", http.StatusBadRequest)
		return
	}
	ctx := req.Context()
	ticker := time.NewTicker(time.Second)
	dials := h.acquireDialer(tunnelID)
	done := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Minute))
			if err != nil {
				conn.Close()
				return
			}
		case <-ctx.Done():
			conn.Close()
			return
		case dials <- dialState{client: conn, done: done}:
			<-done
			return
		}
	}
}

func (h *H) validToken(token string, tunnelID string) error {
	return nil
}

func getToken(req *http.Request) string {
	authtoken := req.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(authtoken, prefix) {
		return ""
	}
	return authtoken[len(prefix):]
}
