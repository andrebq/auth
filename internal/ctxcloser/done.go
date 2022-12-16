package ctxcloser

import (
	"context"
	"io"
)

func WhenDone(ctx context.Context, closer io.Closer) (chan<- struct{}, <-chan error) {
	exit := make(chan struct{})
	leave := make(chan error, 1)
	go func() {
		select {
		case <-ctx.Done():
			leave <- closer.Close()
			close(leave)
		case <-exit:
			close(leave)
		}
	}()
	return exit, leave
}
