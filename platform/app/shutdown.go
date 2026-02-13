package app

import (
	"os"
	"os/signal"
	"syscall"
)

func (a *App) waitShutdown() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
	a.Log.Info().Msg("shutting down...")

	a.GRPC.Stop()
	a.DB.Close()
	a.cancel()
}
