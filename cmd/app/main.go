package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/StewardMcCormick/Paste_Bin/config"
	"github.com/StewardMcCormick/Paste_Bin/internal/controller/HTTP"
	"github.com/StewardMcCormick/Paste_Bin/internal/controller/HTTP/handlers"
	"github.com/StewardMcCormick/Paste_Bin/pkg/httpserver"
	"github.com/StewardMcCormick/Paste_Bin/pkg/logging"
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

func AppRun(ctx context.Context, cfg *config.Config) {
	logger, err := logging.NewLogger(cfg.Logger, cfg.App.Env, cfg.App.Name, cfg.App.Version)
	if err != nil {
		panic(err)
	}

	handler := handlers.NewHandler()
	router := HTTP.NewRouter(handler, logger)

	server := httpserver.New(router, &cfg.Server)

	go func() {
		logger.Info(fmt.Sprintf("Server starts on %s:%s", cfg.Server.Host, cfg.Server.Port))
		err = server.Run()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	<-sig
	logger.Info("[SHUTDOWN] Start shutting down...")

	logger.Info("[SHUTDOWN] Start closing server...")
	err = server.Close()
	if err != nil {
		logger.Error(fmt.Sprintf("[SHUTDOWN] Server closing error: %v", err))
	} else {
		logger.Info("[SHUTDOWN] Server closed")
	}

	err = logger.Sync()
	if err != nil && !errors.Is(err, syscall.ENOTTY) && !errors.Is(err, syscall.EINVAL) && !errors.Is(err, syscall.EBADF) {
		logger.Error(fmt.Sprintf("[SHUTDOWN] Log sync error: %v", err))
	}

	logger.Info("[SHUTDOWN] Shutdown completed")
}
