package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/StewardMcCormick/Paste_Bin/internal/adapter/postgres"
	"github.com/StewardMcCormick/Paste_Bin/internal/adapter/redis"
	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	"github.com/StewardMcCormick/Paste_Bin/internal/handler"
	"github.com/StewardMcCormick/Paste_Bin/internal/handler/middleware"
	pasteH "github.com/StewardMcCormick/Paste_Bin/internal/handler/paste"
	userH "github.com/StewardMcCormick/Paste_Bin/internal/handler/user"
	"github.com/StewardMcCormick/Paste_Bin/internal/repository"
	appcache "github.com/StewardMcCormick/Paste_Bin/internal/repository/cache"
	"github.com/StewardMcCormick/Paste_Bin/internal/repository/paste"
	userUseCase "github.com/StewardMcCormick/Paste_Bin/internal/usecase/auth"
	pasteUseCase "github.com/StewardMcCormick/Paste_Bin/internal/usecase/paste"
	"github.com/StewardMcCormick/Paste_Bin/internal/util/security"
	"github.com/StewardMcCormick/Paste_Bin/internal/util/validation"
	views "github.com/StewardMcCormick/Paste_Bin/internal/util/views_worker"
	"github.com/StewardMcCormick/Paste_Bin/pkg/httpserver"
	"github.com/StewardMcCormick/Paste_Bin/pkg/logging"
	"github.com/StewardMcCormick/Paste_Bin/pkg/migrations"
	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
)

func (a *App) InitLogger(cfg logging.Config) error {
	log, err := logging.NewLogger(cfg, a.cfg.App.Env, a.cfg.App.Name, a.cfg.App.Version)

	if err != nil {
		return fmt.Errorf("[START] logger init error - %w", err)
	}

	log.Info("[START] Logger initialization completed")
	a.log = log
	return nil
}

func (a *App) InitPool(ctx context.Context, cfg postgres.Config) error {
	a.log.Info("[START] PGX pool initialization...")
	pool, err := postgres.NewPool(ctx, &cfg)

	if err != nil {
		return nil
	}

	a.pool = pool
	a.log.Info("[START] PGX initialization completed")

	a.log.Info("[START] DataBase migrations executing...")
	err = migrations.Exec(cfg.DbUrl, cfg.MigrationsPath)
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			a.log.Info("[START] Migrations - nothing to change")
		} else {
			a.log.Error(err.Error())
			return err
		}
	}
	a.log.Info("[START] DataBase migrations executing completed")

	return nil
}

func (a *App) InitRedis(cfg redis.Config) error {
	a.log.Info("[START] Redis client initialization...")
	redisManager, err := redis.NewManager(cfg)
	if err != nil {
		return err
	}
	a.log.Info("[START] Redis client initialization completed")

	a.redis = redisManager
	return nil
}

func (a *App) InitViewsWorker(ctx context.Context) error {
	viewWorker := views.NewViewsWorker(a.pool, 10, 10*time.Millisecond)
	a.log.Info("[START] View Worker started")

	a.viewWorker = viewWorker
	return nil
}

func (a *App) InitServer() error {
	pasteCache := appcache.NewPasteCache(a.redis.GetPasteCacheClient())
	apiKeyCache := appcache.NewAPIKeyCache(a.redis.GetAPIKeyCacheClient())

	uowFactory := repository.NewUWFactory(a.pool, apiKeyCache)
	pasteRepo := paste.NewRepository(a.pool, pasteCache)
	securityUtil := security.NewUtil()

	userValid := validation.NewValidator[*dto.UserRequest](validator.New(validator.WithRequiredStructEnabled()))
	pasteValid := validation.NewValidator[*dto.PasteRequest](validator.New(validator.WithRequiredStructEnabled()))

	authUc := userUseCase.NewUseCase(uowFactory, securityUtil, userValid, a.cfg.Auth)
	pasteUc := pasteUseCase.NewUseCase(a.cfg.Paste, pasteRepo, pasteValid, securityUtil, a.viewWorker)

	a.log.Info("[START] Server initialization...")

	logMid := middleware.NewLogging(a.log)
	recoverMid := middleware.NewRecoverer()
	envMid := middleware.NewEnv(a.cfg.App.Env)
	validMid := middleware.NewJSONValidation()
	authMid := middleware.NewAuth(authUc)

	userHandler := userH.NewHandler(authUc)
	pasteHandler := pasteH.NewHandlers(pasteUc)
	router := handler.NewRouter(
		userHandler,
		pasteHandler,
		logMid,
		recoverMid,
		envMid,
		validMid,
		authMid,
	)
	server := httpserver.New(router, a.cfg.Server)

	a.server = server

	return nil
}
