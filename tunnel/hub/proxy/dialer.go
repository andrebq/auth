package proxy

import (
	"context"
	"net"

	"github.com/andrebq/auth/tunnel/hub"
)

// LocalToRemote takes connections from the given listener and proxies them
// through the given tunnel at wsBase using the provided token.
//
// It only returns when lst.Accept returns an error
func LocalToRemote(ctx context.Context, lst net.Listener, wsBase, token, tunnelID string) error {
	for {
		conn, err := lst.Accept()
		if err != nil {
			return err
		}
		remote, err := hub.Dial(wsBase, token, tunnelID)
		if err != nil {
			conn.Close()
			continue
		}

		go proxyConn(ctx, conn, remote)
	}
}
