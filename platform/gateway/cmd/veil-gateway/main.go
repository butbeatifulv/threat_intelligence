package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/butbeautifulv/veil/platform/gateway/internal/config"
	"github.com/butbeautifulv/veil/platform/gateway/internal/gateway"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck())
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	mux := gateway.NewMux(cfg)
	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	go func() {
		<-ctx.Done()
		shctx, cc := context.WithTimeout(context.Background(), 8*time.Second)
		defer cc()
		_ = srv.Shutdown(shctx)
	}()

	logger := slog.Default()
	logger.Info("veil-gateway listening",
		slog.String("addr", cfg.ListenAddr),
		slog.String("graph_api", cfg.GraphAPIURL.String()),
		slog.String("engage_api", cfg.EngageAPIURL.String()),
		slog.String("graph_mcp", cfg.GraphMCPURL.String()),
		slog.String("engage_mcp", cfg.EngageMCPURL.String()),
	)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func runHealthcheck() int {
	cfg, err := config.Load()
	if err != nil {
		return 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if !gateway.UpstreamsHealthy(ctx, cfg) {
		return 1
	}
	return 0
}
