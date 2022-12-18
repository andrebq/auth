package wsutil

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type (
	Hub struct {
		newChan chan ChanLike
	}

	ChanLike interface {
		Read() <-chan []byte
		Done() <-chan struct{}
		Write() chan<- []byte
	}

	pkt struct {
		src     ChanLike
		content []byte
	}
)

var (
	upgrader = websocket.Upgrader{}
)

func NewHub() *Hub {
	return &Hub{
		newChan: make(chan ChanLike),
	}
}

func (h *Hub) Register(c ChanLike) {
	h.newChan <- c
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ch := NewChan(conn)
	go ch.Loop()
	h.Register(ch)
	select {
	case <-req.Context().Done():
	case <-ch.Done():
	}
	ch.Close()
}

func (h *Hub) Run(ctx context.Context) error {
	broadcast := make(chan pkt, 1000)
	nodes := make(map[ChanLike]signal)
	for {
		select {
		case ch := <-h.newChan:
			go h.handleChan(ctx, ch, broadcast)
			nodes[ch] = signal{}
		case pkt := <-broadcast:
			for k := range nodes {
				if k == pkt.src {
					continue
				}
				println("sending", string(pkt.content), "to", fmt.Sprintf("%p", k))
				select {
				case k.Write() <- pkt.content:
					println("sent", string(pkt.content), "to", fmt.Sprintf("%p", k))
				case <-k.Done():
					delete(nodes, k)
				default:
					println("slow consumer!")
				}
			}
		}
		// clean nodes that might have disconnected
		for k := range nodes {
			select {
			case <-k.Done():
				delete(nodes, k)
			default:
				continue
			}
		}
	}
}

func (h *Hub) handleChan(ctx context.Context, ch ChanLike, bd chan pkt) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-ch.Read():
			println("got", string(data), "from", fmt.Sprintf("%p", ch))
			bd <- pkt{content: data, src: ch}
		}
	}
}
