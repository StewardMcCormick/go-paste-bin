package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/StewardMcCormick/Paste_Bin/config"
	"github.com/StewardMcCormick/Paste_Bin/internal/adapter/postgres"
	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
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

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	AppRun(ctx, cfg)

	cancel()
}

func AppRun(ctx context.Context, cfg *config.Config) {
	logger, err := logging.NewLogger(cfg.Logger, cfg.App.Env, cfg.App.Name, cfg.App.Version)
	if err != nil {
		panic(err)
	}
	logger.Info("[START] Logger initialization completed")

	logger.Info("[START] PGX pool initialization...")
	pool, err := postgres.NewPool(ctx, &cfg.Postgres)
	if err != nil {
		panic(err)
	}
	logger.Info("[START] PGX initialization completed")

	logger.Info("[START] DataBase migrations executing...")
	err = migrations.Exec(cfg.Postgres.DbUrl, cfg.Postgres.MigrationsPath)
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("[START] Migrations - nothing to change")
		} else {
			panic(err)
		}
	}
	logger.Info("[START] DataBase migrations executing completed")

	pasteCache := appcache.NewInMemoryCache[string, *domain.PasteContent](ctx, 10)
	apiKeyCache := appcache.NewInMemoryCache[int64, *domain.APIKey](ctx, 10)

	uowFactory := repository.NewUWFactory(pool, apiKeyCache)
	pasteRepo := paste.NewRepository(pool, pasteCache)
	securityUtil := security.NewUtil()

	userValid := validation.NewValidator[*dto.UserRequest](validator.New(validator.WithRequiredStructEnabled()))
	pasteValid := validation.NewValidator[*dto.PasteRequest](validator.New(validator.WithRequiredStructEnabled()))

	viewWorker := views.NewViewsWorker(pool, 10, 10*time.Millisecond)
	viewWorker.Start(ctx)
	logger.Info("[START] View Worker started")

	authUc := userUseCase.NewUseCase(uowFactory, securityUtil, userValid, cfg.Auth)
	pasteUc := pasteUseCase.NewUseCase(cfg.Paste, pasteRepo, pasteValid, securityUtil, viewWorker)

	logger.Info("[START] Server initialization...")

	logMid := middleware.NewLogging(logger)
	recoverMid := middleware.NewRecoverer()
	envMid := middleware.NewEnv(cfg.App.Env)
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
	server := httpserver.New(router, &cfg.Server)

	go func() {
		logger.Info(fmt.Sprintf("[START] Server starts on %s:%s", cfg.Server.Host, cfg.Server.Port))
		err = server.Run()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	<-sig
	logger.Info("[SHUTDOWN] Start shutting down...")

	viewWorker.Close(ctx)
	logger.Info("[SHUTDOWN] View Worker closed")

	pasteCache.Close(ctx)
	apiKeyCache.Close(ctx)
	logger.Info("[SHUTDOWN] Cache closed")

	pool.Close()
	logger.Info("[SHUTDOWN] PGX close completed")

	err = server.Close()
	if err != nil {
		logger.Panic(fmt.Sprintf("[SHUTDOWN] Server closing error: %v", err))
	}
	logger.Info("[SHUTDOWN] Server close completed")

	err = logger.Sync()
	if err != nil && !errors.Is(err, syscall.ENOTTY) && !errors.Is(err, syscall.EINVAL) && !errors.Is(err, syscall.EBADF) {
		logger.Panic(fmt.Sprintf("[SHUTDOWN] Log sync error: %v", err))
	}
	logger.Info("[SHUTDOWN] Logger sync completed")

	logger.Info("[SHUTDOWN] Shutdown completed")
}
