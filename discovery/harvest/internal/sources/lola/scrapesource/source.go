// Package scrapesource registers the lola scrape source with harvest/internal/factory.
package scrapesource

import (
	"context"
	"os"
	"strings"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/cache"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/factory"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
	lolascrapepub "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/lola/internal/scrapepub"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/lola/internal/usecase"
)

func init() {
	factory.Register("lola", func() factory.Source { return &Source{} })
}

// Source scrapes LOLBAS, GTFOBins, LOFTS, and MITRE ATT&CK STIX.
type Source struct{}

func (s *Source) Name() string { return "lola" }

func (s *Source) Policy() factory.FetchPolicy { return factory.PolicyPeriodic }

func (s *Source) Run(ctx context.Context, deps *factory.ScrapeDeps) error {
	pub, err := deps.Publisher("lola")
	if err != nil {
		return err
	}
	repo := lolascrapepub.NewFromRaw(pub)
	fc := feeds.NewSourceClient(feeds.SourceHTTPConfig{
		CacheDir: cacheDir(),
		Log:      deps.Log,
		Proxy: feeds.ProxyOptions{
			URLsEnv: "LOLA_PROXY_URLS",
			ModeEnv: "LOLA_PROXY_MODE",
			Label:   "lola",
		},
	})
	scraper := usecase.NewScraperUsecase(repo, deps.Log, fc, deps.Ledger)
	return scraper.Run(ctx)
}

func cacheDir() string {
	if v := strings.TrimSpace(os.Getenv("LOLA_CACHE_DIR")); v != "" {
		return v
	}
	return cache.DefaultDir()
}
