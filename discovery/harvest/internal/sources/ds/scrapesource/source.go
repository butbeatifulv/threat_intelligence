// Package scrapesource registers the ds scrape source with harvest/internal/factory.
package scrapesource

import (
	"context"
	"os"
	"strings"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/cache"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/factory"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
	dsscrapepub "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/ds/internal/scrapepub"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/ds/internal/usecase"
)

func init() {
	factory.Register("ds", func() factory.Source { return &Source{} })
}

// Source scrapes Sigma, YARA, Atomic Red Team, and Caldera via GitHub.
type Source struct{}

func (s *Source) Name() string { return "ds" }

func (s *Source) Policy() factory.FetchPolicy { return factory.PolicyPeriodic }

func (s *Source) Run(ctx context.Context, deps *factory.ScrapeDeps) error {
	pub, err := deps.Publisher("ds")
	if err != nil {
		return err
	}
	repo := dsscrapepub.NewFromRaw(pub)
	fc := feeds.NewSourceClient(feeds.SourceHTTPConfig{
		CacheDir: cacheDir(),
		Log:      deps.Log,
		Proxy: feeds.ProxyOptions{
			URLsEnv: "DS_PROXY_URLS",
			ModeEnv: "DS_PROXY_MODE",
			Label:   "ds",
		},
	})
	ing := usecase.NewIngestor(repo, deps.Log, fc, deps.Ledger)
	return ing.Run(ctx)
}

func cacheDir() string {
	if v := strings.TrimSpace(os.Getenv("DS_CACHE_DIR")); v != "" {
		return v
	}
	return cache.DefaultDir()
}
