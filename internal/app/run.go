package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"syscall"
)

func (a *App) Run(ctx context.Context) {
	a.viewWorker.Start(ctx)
	a.dbCleanUpWorker.Start(ctx)
	a.log.Info("[START] View Worker started")

	go func() {
		a.log.Info(fmt.Sprintf("[START] Server starts on %s:%s", a.cfg.Server.Host, a.cfg.Server.Port))
		err := a.server.Run()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
}

func (a *App) Shutdown(ctx context.Context) {
	a.log.Info("[SHUTDOWN] Start shutting down...")

	a.viewWorker.Stop(ctx)
	a.log.Info("[SHUTDOWN] View Worker closed")

	a.dbCleanUpWorker.Stop(ctx)
	a.log.Info("[SHUTDOWN] DB Clean-Up Worker closed")

	err := a.redis.Close()
	if err != nil {
		a.log.Error(fmt.Sprintf("[SHUTDOWN] Redis client close error - %v", err))
	} else {
		a.log.Info("[SHUTDOWN] Redis client closed")
	}

	a.pool.Close()
	a.log.Info("[SHUTDOWN] PGX close completed")

	err = a.server.Close()
	if err != nil {
		log.Panicf("[SHUTDOWN] Server closing error: %v", err)
	}
	a.log.Info("[SHUTDOWN] Server close completed")

	err = a.log.Sync()
	if err != nil && !errors.Is(err, syscall.ENOTTY) && !errors.Is(err, syscall.EINVAL) && !errors.Is(err, syscall.EBADF) {
		log.Panicf("[SHUTDOWN] Log sync error: %v", err)
	}
	a.log.Info("[SHUTDOWN] Logger sync completed")

	a.log.Info("[SHUTDOWN] Shutdown completed")
}
