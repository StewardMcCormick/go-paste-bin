package main

import (
	"context"
	"github.com/StewardMcCormick/Paste_Bin/config"
	"github.com/StewardMcCormick/Paste_Bin/internal/http/controller"
	"github.com/StewardMcCormick/Paste_Bin/pkg/httpserver"
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
	router := controller.Router()

	server := httpserver.New(router, &cfg.Server)
	err := server.Run()
	if err != nil {
		panic(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
	server.Close()
}
