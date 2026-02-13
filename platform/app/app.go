package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/config"
	grpcServer "github.com/dandirahmadani19/distributed-saga-orchestrator/platform/grpc"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/logger"
	"github.com/dandirahmadani19/distributed-saga-orchestrator/platform/postgres"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

type App struct {
	Cfg    *config.Config
	Log    *logger.Logger
	DB     *sql.DB
	GRPC   *grpcServer.Server
	ctx    context.Context
	cancel context.CancelFunc

	grpcFactory GRPCOptionFactory
}

func New(opts ...Option) (*App, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	a := &App{}

	for _, opt := range opts {
		opt(a)
	}

	log := logger.New(cfg.App.Name)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := postgres.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("database connected to " + cfg.Postgres.DBName)

	var grpcOpts []grpc.ServerOption
	if a.grpcFactory != nil {
		grpcOpts = a.grpcFactory(cfg, log)
	}

	g, err := grpcServer.New(fmt.Sprintf("%d", cfg.GRPC.Port), grpcOpts...)
	if err != nil {
		return nil, err
	}

	if cfg.GRPC.Reflection {
		grpcServer.EnableReflection(g)
		log.Info().Msg("gRPC reflection enabled")
	}

	return &App{
		Cfg:    cfg,
		Log:    log,
		DB:     db,
		GRPC:   g,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}
