package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/butbeautifulv/veil/pipeline/ned/internal/components"
	"github.com/butbeautifulv/veil/pkg/observability"
)

func main() {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	obsCfg := observability.LoadConfigFromEnv("veil-pipeline-worker")
	shutdown, err := observability.Init(rootCtx, obsCfg)
	if err != nil {
		slog.Error("otel init", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer func() {
		shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = shutdown(shCtx)
	}()

	log := observability.NewLogger("", "veil-pipeline-worker", os.Stdout)
	go func() {
		if err := <-observability.StartMetricsServer(rootCtx, obsCfg.MetricsListen, log); err != nil {
			log.Warn("metrics server failed", slog.String("err", err.Error()))
		}
	}()

	if err := components.Run(rootCtx, log); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("exit", slog.String("err", err.Error()))
		os.Exit(1)
	}
	log.Info("shutdown complete")
}
