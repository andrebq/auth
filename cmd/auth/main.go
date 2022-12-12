package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/andrebq/auth/cmd/auth/cmdlib"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app := cmdlib.NewApp(os.Stdin)
	err := app.RunContext(ctx, os.Args)
	if errors.Is(err, context.Canceled) {
		err = nil
	}
	if err != nil {
		log.Error().Err(err).Msg("Application failed")
	}
}
