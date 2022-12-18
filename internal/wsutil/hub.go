package wsutil

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type (
	Hub struct {
		newChan chan ChanLike
	}

	ChanLike interface {
		ID() [16]byte
		Read() <-chan []byte
		Done() <-chan struct{}
		Write() chan<- []byte
	}

	pkt struct {
		src     ChanLike
		to      [16]byte
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

func (h *Hub) Register(c ChanLike, endpoint [16]byte) {
	h.newChan <- c
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	endpointID, err := uuid.Parse(req.FormValue("endpoint_id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ch := NewChan(conn, endpointID)
	go ch.Loop()
	h.Register(ch, endpointID)
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
				if k == pkt.src || k.ID() != pkt.to {
					continue
				}
				select {
				case k.Write() <- pkt.content:
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
			if len(data) < 16 {
				continue
			}
			p := pkt{content: data, src: ch}
			copy(p.to[:], data[:16])
			bd <- p
		}
	}
}
