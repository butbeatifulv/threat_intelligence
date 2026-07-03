package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	connats "github.com/butbeautifulv/veil/pipeline/connector/nats"
	"github.com/butbeautifulv/veil/pkg/observability"
)

func main() {
	obsCfg := observability.LoadConfigFromEnv("veil-engage-events-worker")
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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

	logger := observability.NewLogger("", "veil-engage-events-worker", os.Stdout)
	observability.StartMetricsServer(rootCtx, obsCfg.MetricsListen, logger)

	natsURL := env("NATS_URL", "nats://127.0.0.1:4222")
	filter := env("ENGAGE_EVENTS_FILTER", "engage.events.>")
	ingestRun := env("INGEST_SUBJECT", "ingest.engage.tool_run")
	ingestFinding := env("INGEST_FINDING_SUBJECT", "ingest.engage.finding")

	logger.Info("engage-events worker starting", slog.String("nats", natsURL), slog.String("filter", filter))
	if err := connats.RunEngageEventsConsumer(rootCtx, logger, natsURL, filter, ingestRun, ingestFinding); err != nil && rootCtx.Err() == nil {
		logger.Error("consumer stopped", slog.Any("err", err))
		os.Exit(1)
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
