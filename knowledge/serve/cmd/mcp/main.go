package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/butbeautifulv/veil/knowledge/serve/internal/components"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/config"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/transport/securityhttp"
	"github.com/butbeautifulv/veil/pkg/observability"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck())
	}

	var (
		env       = flag.String("env", envOr("MCP_ENV", "local"), "local|dev|prod")
		neo4jURI  = flag.String("neo4j-uri", envOr("NEO4J_URI", "neo4j://localhost:7687"), "neo4j uri")
		neo4jUser = flag.String("neo4j-user", envOr("NEO4J_USER", "neo4j"), "neo4j username")
		neo4jPass = flag.String("neo4j-pass", envOr("NEO4J_PASS", "neo4jpassword"), "neo4j password")
		neo4jDB   = flag.String("neo4j-db", envOr("NEO4J_DB", "neo4j"), "neo4j database")
	)
	flag.Parse()

	obsCfg := observability.LoadConfigFromEnv("veil-mcp")
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

	logger := components.SetupMCPLogger(*env)
	cfg := config.LoadMCP()
	cfg.Neo4j = config.Neo4jConfig{
		URI:      *neo4jURI,
		Username: *neo4jUser,
		Password: *neo4jPass,
		Database: *neo4jDB,
	}
	cfg.Env = *env
	cfg.Security = config.LoadSecurityForEnv(*env)

	if err := config.ValidateSecurity(cfg.Security, cfg.Auth.Enabled); err != nil {
		log.Fatal(err)
	}

	c, err := components.InitMCP(cfg, logger)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Shutdown()

	var httpSrv *http.Server
	if cfg.MCPHTTP.Enabled {
		addr := cfg.MCPHTTP.ResolveListen()
		rh, rt, wt, idle := securityhttp.HTTPServerTimeouts()
		httpSrv = &http.Server{
			Addr:              addr,
			Handler:           observability.WrapHandler("veil-mcp", c.MCPHTTPHandler(cfg)),
			ReadHeaderTimeout: time.Duration(rh) * time.Second,
			ReadTimeout:       time.Duration(rt) * time.Second,
			WriteTimeout:      time.Duration(wt) * time.Second,
			IdleTimeout:       time.Duration(idle) * time.Second,
		}
		go func() {
			logger.Info("mcp http listening",
				slog.String("addr", addr),
				slog.String("path", cfg.MCPHTTP.Path),
			)
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("mcp http server stopped", slog.Any("err", err))
			}
		}()
		go func() {
			<-rootCtx.Done()
			shctx, cc := context.WithTimeout(context.Background(), 8*time.Second)
			defer cc()
			if httpSrv != nil {
				_ = httpSrv.Shutdown(shctx)
			}
		}()
	}

	if cfg.MCPHTTP.Enabled {
		<-rootCtx.Done()
		return
	}

	if err := c.MCPServer.Run(rootCtx, os.Stdin, os.Stdout); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		logger.Error("mcp server stopped", slog.Any("err", err))
		time.Sleep(200 * time.Millisecond)
		os.Exit(1)
	}
}

func runHealthcheck() int {
	cfg := config.LoadMCP()
	c, err := components.InitMCP(cfg, components.SetupMCPLogger(cfg.Env))
	if err != nil {
		return 1
	}
	defer c.Shutdown()
	if err := c.Read.Ping(context.Background()); err != nil {
		return 1
	}
	return 0
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
