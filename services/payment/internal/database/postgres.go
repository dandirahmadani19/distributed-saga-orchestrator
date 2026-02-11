package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/services/payment/internal/config"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

func Postgres(cfg *config.DatabaseConfig) (*sql.DB, error) {
	return newPostgres(cfg)
}

func newPostgres(cfg *config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().Msg(fmt.Sprintf("✅ Success to connect database: %s", cfg.DBName))
	return db, nil
}
