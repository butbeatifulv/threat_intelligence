// Package scrapesource registers the vuln scrape source with harvest/internal/factory.
package scrapesource

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/cache"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/factory"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
	vulnscrapepub "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/vuln/internal/scrapepub"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/vuln/internal/config"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/vuln/internal/usecase"
)

func init() {
	factory.Register("vuln", func() factory.Source { return &Source{} })
}

// Source scrapes NVD, Metasploit, Exploit-DB, and optional Vulners.
type Source struct{}

func (s *Source) Name() string { return "vuln" }

func (s *Source) Policy() factory.FetchPolicy { return factory.PolicyPeriodic }

func (s *Source) Run(ctx context.Context, deps *factory.ScrapeDeps) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	pub, err := deps.Publisher("vuln")
	if err != nil {
		return err
	}
	repo := vulnscrapepub.NewFromRaw(pub)
	fc := feeds.NewSourceClient(feeds.SourceHTTPConfig{
		CacheDir: vulnCacheDir(),
		Log:      deps.Log,
		Proxy: feeds.ProxyOptions{
			URLsEnv: "VULN_PROXY_URLS",
			ModeEnv: "VULN_PROXY_MODE",
			Label:   "vuln",
		},
		Timeout: 60 * time.Second,
	})
	scraper := usecase.NewScraperUsecase(repo, deps.Log, cfg.NVD.APIKey, fc, deps.Ledger)
	return scraper.Run(ctx)
}

func vulnCacheDir() string {
	if v := strings.TrimSpace(os.Getenv("VULN_CACHE_DIR")); v != "" {
		return v
	}
	return cache.DefaultDir()
}
