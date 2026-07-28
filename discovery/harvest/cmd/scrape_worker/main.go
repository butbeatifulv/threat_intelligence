package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/factory"
	"github.com/butbeautifulv/veil/pkg/observability"
	"go.opentelemetry.io/otel/attribute"

	_ "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/coderules/scrapesource"
	_ "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/ds/scrapesource"
	_ "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/lola/scrapesource"
	_ "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/nuclei/scrapesource"
	_ "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/sbom/scrapesource"
	_ "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/ti/scrapesource"
	_ "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/vuln/scrapesource"
)

func main() {
	obsCfg := observability.LoadConfigFromEnv("veil-scrape-worker")
	ctx := context.Background()
	shutdown, err := observability.Init(ctx, obsCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = shutdown(shCtx)
	}()

	logger := observability.NewLogger("", "veil-scrape-worker", os.Stdout)
	names := factory.ParseSourceNames(os.Getenv("SCRAPE_SOURCES"))

	runCtx, span := observability.StartWorkerSpan(ctx, "scrape", "scrape.run",
		attribute.StringSlice("scrape.sources", names),
	)
	defer span.End()

	logger.Info("scrape_worker starting", slog.Any("sources", names))
	if err := factory.Run(runCtx, factory.RunOptions{
		SourceNames: names,
		Log:         logger,
	}); err != nil {
		observability.RecordError(span, err)
		log.Fatal(err)
	}
}
