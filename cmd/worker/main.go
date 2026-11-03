package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dealngon/backend/internal/bootstrap"
	"github.com/dealngon/backend/internal/config"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := bootstrap.New(ctx, cfg)
	if err != nil {
		slog.Error("startup", "error", err.Error())
		os.Exit(1)
	}
	defer app.Close()

	run := func() {
		if err := app.Sync.Run(ctx); err != nil {
			slog.Error("sync", "error", err.Error())
		} else {
			slog.Info("sync complete")
		}
	}
	run()
	ticker := time.NewTicker(cfg.WorkerInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}
