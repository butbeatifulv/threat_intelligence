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

	"github.com/butbeautifulv/veil/knowledge/serve/internal/components"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/config"
	authmw "github.com/butbeautifulv/veil/knowledge/serve/internal/auth/middleware"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/transport/httpserver"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/transport/securityhttp"
	"github.com/butbeautifulv/veil/pkg/observability"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck())
	}

	cfg := config.LoadAPI()
	obsCfg := observability.LoadConfigFromEnv("veil-api")

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdown, err := observability.Init(rootCtx, obsCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = shutdown(shCtx)
	}()

	logger := components.SetupLogger(cfg.Env)
	httpserver.SetProdMode(cfg.Security.Prod)

	c, err := components.InitAPI(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Shutdown()

	mux := http.NewServeMux()
	httpserver.Register(mux, c.Read, c.Playbook, c.Procedure, c.Framework)
	observability.RegisterMetrics(mux)
	inner := observability.ChainHTTP("veil-api", mux)
	handler := securityhttp.Harden(cfg.Security, cfg.Security.APIBodyLimit,
		authmw.Auth(c.Auth, false, cfg.Security, inner))

	rh, rt, wt, idle := securityhttp.HTTPServerTimeouts()
	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           handler,
		ReadHeaderTimeout: time.Duration(rh) * time.Second,
		ReadTimeout:       time.Duration(rt) * time.Second,
		WriteTimeout:      time.Duration(wt) * time.Second,
		IdleTimeout:       time.Duration(idle) * time.Second,
	}

	go func() {
		<-rootCtx.Done()
		shctx, cc := context.WithTimeout(context.Background(), 8*time.Second)
		defer cc()
		_ = srv.Shutdown(shctx)
	}()

	logger.Info("api listening", slog.String("addr", cfg.ListenAddr))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func runHealthcheck() int {
	cfg := config.LoadAPI()
	c, err := components.InitAPI(cfg)
	if err != nil {
		return 1
	}
	defer c.Shutdown()
	if err := c.Read.Ping(context.Background()); err != nil {
		return 1
	}
	return 0
}
