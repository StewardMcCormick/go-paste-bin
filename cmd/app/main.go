package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/StewardMcCormick/Paste_Bin/config"
	"github.com/StewardMcCormick/Paste_Bin/internal/adapter/postgres"
	"github.com/StewardMcCormick/Paste_Bin/internal/handler"
	"github.com/StewardMcCormick/Paste_Bin/internal/handler/middleware"
	userH "github.com/StewardMcCormick/Paste_Bin/internal/handler/user"
	userRepo "github.com/StewardMcCormick/Paste_Bin/internal/repository/user"
	userUseCase "github.com/StewardMcCormick/Paste_Bin/internal/usecase/user"
	"github.com/StewardMcCormick/Paste_Bin/internal/util/security"
	"github.com/StewardMcCormick/Paste_Bin/pkg/httpserver"
	"github.com/StewardMcCormick/Paste_Bin/pkg/logging"
	"github.com/StewardMcCormick/Paste_Bin/pkg/migrations"
	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic(err)
	}

	AppRun(context.Background(), cfg)
}

// logging in -> ... TODO

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

	userRepository := userRepo.NewRepository(pool)
	securityUtil := security.NewUtil()
	valid := validator.New(validator.WithRequiredStructEnabled())
	userUC := userUseCase.NewUseCase(userRepository, securityUtil, valid, cfg.Auth)

	logger.Info("[START] Server initialization...")

	logMid := middleware.NewLogging(logger)
	recoverMid := middleware.NewRecoverer()
	envMid := middleware.NewEnv(cfg.App.Env)
	validMid := middleware.NewJSONValidation()

	userHandler := userH.NewHandler(userUC)
	router := handler.NewRouter(
		userHandler,
		logMid,
		recoverMid,
		envMid,
		validMid,
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
