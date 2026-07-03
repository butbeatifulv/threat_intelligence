package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/butbeautifulv/veil/knowledge/ingest/internal/components"
	"github.com/butbeautifulv/veil/knowledge/ingest/internal/config"
	ingestloop "github.com/butbeautifulv/veil/knowledge/ingest/internal/ingest"
	"github.com/butbeautifulv/veil/pkg/natsjet"
	"github.com/butbeautifulv/veil/pkg/observability"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
)

func main() {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	obsCfg := observability.LoadConfigFromEnv("veil-ingest-worker")
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

	log := observability.NewLogger("", "veil-ingest-worker", os.Stdout)
	observability.StartMetricsServer(rootCtx, obsCfg.MetricsListen, log)

	if err := run(rootCtx, log); err != nil && !errors.Is(err, context.Canceled) {
		log.Error("exit", slog.String("err", err.Error()))
		os.Exit(1)
	}
	log.Info("shutdown complete")
}

func run(rootCtx context.Context, log *slog.Logger) error {
	cfg := config.Load()

	rt, err := components.Init(rootCtx, cfg, log)
	if err != nil {
		return err
	}
	defer rt.Shutdown(rootCtx)

	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return err
	}
	defer nc.Drain()

	js, err := nc.JetStream()
	if err != nil {
		return err
	}
	if err := natsjet.EnsureIngestStream(js); err != nil {
		return err
	}

	sub, err := js.PullSubscribe(cfg.Subject, cfg.Durable, nats.BindStream(cfg.Stream))
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	log.Info("ingest-worker started",
		slog.String("nats", cfg.NATSURL),
		slog.String("stream", cfg.Stream),
		slog.String("durable", cfg.Durable),
		slog.String("subject", cfg.Subject),
	)

	eg, ctx := errgroup.WithContext(rootCtx)
	eg.Go(func() error {
		return ingestloop.RunPullLoop(ctx, log, sub, cfg.Batch, cfg.MaxWait, rt)
	})
	return eg.Wait()
}
