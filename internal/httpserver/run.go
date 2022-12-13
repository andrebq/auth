package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func Run(ctx context.Context, addr string, port uint, handler http.Handler) error {
	log := log.Logger.With().Str("addr", addr).Uint("port", port).Logger()
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

	log.Info().Str("addr", addr).Uint("port", port).Msg("Starting server")
	err := srv.ListenAndServe()
	cancel()
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	<-shutdownComplete
	return err
}
