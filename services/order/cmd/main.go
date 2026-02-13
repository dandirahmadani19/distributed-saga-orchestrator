package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/logger"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/config"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/database"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/order/internal/server"
	"github.com/rs/zerolog/log"
)

func main() {
	logger := logger.New("order-service")
	logger.Info().Msg("Starting order service")

	// Load configuration
	if err := config.Init(); err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize database
	db, err := database.Postgres(&config.Get().Database)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}

	// Create server
	srv, err := server.New(db, logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create server")
	}

	// Start server
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	srv.GracefulStop()
}
