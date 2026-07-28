package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/butbeautifulv/veil/platform/mcp-gateway/internal/aggregator"
	"github.com/butbeautifulv/veil/platform/mcp-gateway/internal/authstack"
	"github.com/butbeautifulv/veil/platform/mcp-gateway/internal/backend"
	"github.com/butbeautifulv/veil/platform/mcp-gateway/internal/config"
	"github.com/butbeautifulv/veil/platform/mcp-gateway/internal/transport"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	stack, err := authstack.New(ctx, cfg.Auth)
	if err != nil {
		logger.Error("auth init failed", "err", err)
		os.Exit(1)
	}

	httpClient := cfg.HTTPClient()
	agg := aggregator.New(
		&backend.Client{Name: "graph", URL: cfg.GraphMCPURL, HTTP: httpClient},
		&backend.Client{Name: "engage", URL: cfg.EngageMCPURL, HTTP: httpClient},
		stack,
		logger,
	)

	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           transport.Handler(agg, cfg, stack),
		ReadHeaderTimeout: 10 * time.Second,
	}
	logger.Info("unified mcp http listening",
		"listen", cfg.Listen,
		"path", cfg.Path,
		"graph", cfg.GraphMCPURL,
		"engage", cfg.EngageMCPURL,
	)

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("listen failed", "err", err)
		os.Exit(1)
	}
}
