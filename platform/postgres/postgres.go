package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/config"
	_ "github.com/lib/pq"
)

func NewPostgres(ctx context.Context, cfg config.PostgresConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctxTimeout); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
