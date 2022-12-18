package wsutil

import (
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type (
	// Chan wraps a websocket connection as a go channel, it maintains
	// a connection open by answering Ping messages sent by clients
	Chan struct {
		close sync.Once
		stop  chan signal
		done  chan struct{}
		ping  chan signal
		ws    *websocket.Conn

		downstream chan []byte
		upstream   chan []byte
	}

	signal struct{}
)

func NewChan(ws *websocket.Conn) *Chan {
	ch := Chan{
		stop:       make(chan signal),
		done:       make(chan struct{}),
		ping:       make(chan signal),
		ws:         ws,
		downstream: make(chan []byte),
		upstream:   make(chan []byte),
	}
	return &ch
}

func (c *Chan) Close() error {
	c.close.Do(func() { close(c.stop) })
	<-c.done
	return nil
}

func (c *Chan) Done() <-chan struct{} {
	return c.done
}

func (c *Chan) Read() <-chan []byte {
	return c.downstream
}

func (c *Chan) Write() chan<- []byte {
	return c.upstream
}

func (c *Chan) Ping() error {
	select {
	case <-c.done:
		return errors.New("wsutil:chan: closed")
	case <-c.stop:
		return errors.New("wsutil:chan: closed")
	case c.ping <- signal{}:
		return nil
	}
}

func (c *Chan) KeepAlive(heartbeat time.Duration) {
	t := time.NewTicker(heartbeat)
	defer t.Stop()
	for range t.C {
		c.Ping()
	}
}

func (c *Chan) Loop() {
	defer func() {
		close(c.done)
		c.ws.Close()
	}()
	upstreamErr := make(chan error, 1)
	upstream := func() {
		defer close(upstreamErr)
		for {
			select {
			case <-c.ping:
				err := c.ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second*10))
				if err != nil {
					upstreamErr <- err
					return
				}
			case data, open := <-c.upstream:
				if !open {
					upstreamErr <- errors.New("wsutil:chan: upstream closed")
					return
				}
				c.ws.SetWriteDeadline(time.Now().Add(time.Second * 10))
				err := c.ws.WriteMessage(websocket.BinaryMessage, data)
				if err != nil {
					upstreamErr <- err
					return
				}
			case <-c.stop:
				return
			case <-c.done:
				return
			}
		}
	}
	downstreamErr := make(chan error, 1)
	downstream := func() {
		defer close(downstreamErr)
		for {
			c.ws.SetReadDeadline(time.Now().Add(time.Minute * 5))
			mt, buf, err := c.ws.ReadMessage()
			if err != nil {
				downstreamErr <- err
				return
			}
			if mt == websocket.PongMessage || mt == websocket.PingMessage {
				continue
			}
			select {
			case c.downstream <- buf:
				continue
			case <-c.done:
				return
			case <-c.stop:
				return
			}
		}
	}
	go upstream()
	go downstream()
	for {
		select {
		case <-c.stop:
			return
		case err, open := <-downstreamErr:
			if !open {
				downstreamErr = nil
			}
			_ = err
			return
		case err, open := <-upstreamErr:
			if !open {
				upstreamErr = nil
			}
			_ = err
		}
	}
}
