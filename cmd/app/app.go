package main

import (
	"context"

	"github.com/StewardMcCormick/Paste_Bin/config"
	"github.com/StewardMcCormick/Paste_Bin/internal/adapter/redis"
	views "github.com/StewardMcCormick/Paste_Bin/internal/util/views_worker"
	"github.com/StewardMcCormick/Paste_Bin/pkg/httpserver"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	cfg        *config.Config
	log        *zap.Logger
	pool       *pgxpool.Pool
	server     *httpserver.Server
	redis      *redis.Manager
	viewWorker *views.ViewWorker
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	app := &App{cfg: cfg}

	if err := app.InitLogger(cfg.Logger); err != nil {
		return nil, err
	}

	if err := app.InitPool(context.Background(), cfg.Postgres); err != nil {
		return nil, err
	}

	if err := app.InitRedis(cfg.Redis); err != nil {
		return nil, err
	}

	if err := app.InitViewsWorker(context.Background()); err != nil {
		return nil, err
	}

	if err := app.InitServer(); err != nil {
		return nil, err
	}

	return app, nil
}
