package proxy

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/andrebq/auth/tunnel/hub"
)

// RemoteToLocal opens a new tunnel on wsBase using the given tunnelID
// and tokens to authenticate.
//
// Then, it will accept connections on the given tunnel in a loop, for each
// new connection from the tunnel it will use dialer to acquire a connection
// to a service running on the localhost.
//
// Once that connection is obtained, it will start copying data between both
// connections.
//
// It will only stop once hub.Accept fails
func RemoteToLocal(ctx context.Context, wsBase, token, tunnelID string, dialer func(context.Context) (net.Conn, error)) error {
	for {
		conn, err := hub.Accept(wsBase, token, tunnelID)
		if err != nil {
			return err
		}
		localConn, err := dialer(ctx)
		if err != nil {
			conn.Close()
			// TODO: handle error here
			continue
		}
		go proxyConn(ctx, conn, localConn)
	}
}

func proxyConn(ctx context.Context, a, b net.Conn) {
	ctx, cancel := context.WithCancel(ctx)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(a, b)
	}()
	go func() {
		defer wg.Done()
		io.Copy(b, a)
	}()
	go func() {
		wg.Wait()
		cancel()
	}()
	<-ctx.Done()
	a.Close()
	b.Close()
	wg.Wait()
}
