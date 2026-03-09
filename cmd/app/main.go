package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/StewardMcCormick/Paste_Bin/config"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := NewApp(ctx, cfg)
	if err != nil {
		panic(err)
	}

	app.Run(ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	<-sig
	app.Shutdown(ctx)
}
