package hub

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/gorilla/websocket"
)

type (
	tunnelConn struct {
		ws        *websocket.Conn
		dialer    bool
		tunnelID  string
		reader    io.Reader
		readerErr error
	}

	tunnelAddr struct {
		*tunnelConn
	}
)

var (
	signalPacket = []byte("GREENLIGHT")
)

func Dial(ws, token, tunnelID string) (net.Conn, error) {
	return dial(ws, token, tunnelID, false)
}

func Accept(ws, token, tunnelID string) (net.Conn, error) {
	return dial(ws, token, tunnelID, true)
}

func dial(wsBase, token, tunnelID string, listener bool) (net.Conn, error) {
	wsURL, err := url.Parse(wsBase)
	if err != nil {
		return nil, err
	}
	if listener {
		wsURL.Path = path.Join(wsURL.Path, "ws", "listen")
	} else {
		wsURL.Path = path.Join(wsURL.Path, "ws", "dial")
	}
	values := wsURL.Query()
	values.Add("tunnelId", tunnelID)
	wsURL.RawQuery = values.Encode()
	headers := http.Header{}
	headers.Add("Authorization", fmt.Sprintf("Bearer %v", token))
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), headers)
	if err != nil {
		return nil, err
	}
	println("waiting for signal", "is listener?", listener)
	err = waitForSignal(wsConn)
	if err != nil {
		return nil, err
	}
	return &tunnelConn{tunnelID: tunnelID, ws: wsConn, dialer: false}, nil
}

func (tc *tunnelConn) SetDeadline(dl time.Time) error {
	err := tc.ws.SetReadDeadline(dl)
	if err != nil {
		return err
	}
	return tc.ws.SetWriteDeadline(dl)
}

func (tc *tunnelConn) Write(buf []byte) (int, error) {
	wc, err := tc.ws.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}
	n, err := wc.Write(buf)
	wc.Close()
	if len(buf) != n {
		return n, io.ErrShortWrite
	}
	return n, err
}

func (tc *tunnelConn) SetReadDeadline(dl time.Time) error  { return tc.ws.SetReadDeadline(dl) }
func (tc *tunnelConn) SetWriteDeadline(dl time.Time) error { return tc.ws.SetWriteDeadline(dl) }

func (tc *tunnelConn) RemoteAddr() net.Addr {
	return tunnelAddr{tc}
}

func (tc *tunnelConn) LocalAddr() net.Addr {
	return tunnelAddr{tc}
}

func (tc *tunnelConn) Close() error {
	return tc.ws.Close()
}

func (tc *tunnelConn) Read(out []byte) (int, error) {
	if tc.readerErr != nil {
		return 0, tc.readerErr
	}
	if tc.reader != nil {
		n, err := tc.consumeReader(out)
		if n > 0 || err != nil {
			// either we read something
			// or we got an error other than io.EOF (consumeReader)
			// does not leak io.EOFs to callers
			return n, err
		}
		// at this point, it is safe to say that we finished reading
		// the previous message and should instead
		// acquire a new reader
	}
	// there is no reader
	_, tc.reader, tc.readerErr = tc.ws.NextReader()
	if tc.readerErr != nil {
		return 0, tc.readerErr
	}
	n, err := tc.consumeReader(out)
	if n == 0 && err == nil {
		// we just opened a new reader
		// but it did not return any data
		// this most likely means there is no data to be read
		// therefore we should indicate eof
		return n, io.EOF
	}
	// return whatever consumeReader returned
	return n, err
}

func (tc *tunnelConn) consumeReader(out []byte) (int, error) {
	n, err := tc.reader.Read(out)
	if err == nil {
		// happy path
		return n, nil
	}
	if !errors.Is(err, io.EOF) {
		// an error other than EOF was received
		// therefore we should not continue
		// and we should discard this reader
		tc.reader = nil
		return n, err
	}
	if errors.Is(err, io.EOF) && n > 0 {
		// we reached the end of the previous call
		// but we got something in the buffer
		// clear the reader but do not return io.EOF
		// because there might be more messages
		tc.reader = nil
		return n, nil
	}
	return 0, nil
}

func (ta tunnelAddr) Network() string {
	return "websocket"
}

func (ta tunnelAddr) String() string {
	if ta.dialer {
		return fmt.Sprintf("dial:%v:%v:%v", ta.tunnelID, ta.ws.RemoteAddr(), ta.ws.LocalAddr())
	}
	return fmt.Sprintf("listen:%v:%v:%v", ta.tunnelID, ta.ws.RemoteAddr(), ta.ws.LocalAddr())
}

func waitForSignal(conn *websocket.Conn) error {
	for {
		err := conn.SetReadDeadline(time.Now().Add(time.Minute))
		if err != nil {
			return err
		}
		mt, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		if mt == websocket.PingMessage {
			// DO I NEED TO SEND A PONG?
			continue
		}
		if bytes.Equal(data, signalPacket) {
			return nil
		}
	}
}
