package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// RunProxy starts the HTTP server that proxy authenticated requests to their backend
// when internetFacing is true, this proxy has very strict timeouts and limits,
// otherwise it will not have any limits and such protections should be handled by
// a reverse-proxy in front of auth and by downstream services that should close the
// connection.
func RunProxy(ctx context.Context, internetFacing bool, addr string, port uint, handler http.Handler) error {
	log := log.Logger.With().Str("addr", addr).Uint("port", port).Str("kind", "proxy").Logger()
	srv := http.Server{
		Addr:              fmt.Sprintf("%v:%v", addr, port),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: time.Second,
		MaxHeaderBytes:    10_000,
		Handler:           handler,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}
	if !internetFacing {
		srv.ReadTimeout = 0
		srv.WriteTimeout = 0
		srv.ReadHeaderTimeout = 0
		srv.MaxHeaderBytes = 0
	}
	return runServe(ctx, log, &srv)
}

func Run(ctx context.Context, addr string, port uint, handler http.Handler) error {
	log := log.Logger.With().Str("addr", addr).Uint("port", port).Str("kind", "api").Logger()
	srv := http.Server{
		Addr:              fmt.Sprintf("%v:%v", addr, port),
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		ReadHeaderTimeout: time.Second,
		MaxHeaderBytes:    1_000,
		Handler:           handler,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}
	return runServe(ctx, log, &srv)
}

func runServe(ctx context.Context, log zerolog.Logger, srv *http.Server) error {
	ctx, cancel := context.WithCancel(ctx)
	shutdownComplete := make(chan struct{})

	go func() {
		defer close(shutdownComplete)
		<-ctx.Done()
		log.Info().Msg("Starting server shutdown")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		srv.Shutdown(shutdownCtx)
		cancel()
	}()

	log.Info().Msg("Starting server")
	err := srv.ListenAndServe()
	cancel()
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	<-shutdownComplete
	return err
}
