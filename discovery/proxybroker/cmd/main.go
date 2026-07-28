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

	"proxybroker/internal/broker"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg := broker.Config{
		Addr:        envOr("PROXYBROKER_ADDR", ":8099"),
		Token:       os.Getenv("PROXYBROKER_TOKEN"),
		CheckEvery:  envDur("PROXYBROKER_CHECK_EVERY", 45*time.Second),
		HTTPTimeout: envDur("PROXYBROKER_HTTP_TIMEOUT", 10*time.Second),
		TestURL:     envOr("PROXYBROKER_TEST_URL", "https://example.com/"),
	}

	b := broker.New(cfg, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go b.RunHealthLoop(ctx)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           b.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigQuit
		cancel()
		cctx, ccancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer ccancel()
		_ = srv.Shutdown(cctx)
	}()

	logger.Info("proxybroker listening", slog.String("addr", cfg.Addr))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func envDur(k string, d time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if x, err := time.ParseDuration(v); err == nil && x > 0 {
			return x
		}
	}
	return d
}

