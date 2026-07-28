// Package scrapesource registers the nuclei scrape source with harvest/internal/factory.
package scrapesource

import (
	"context"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/factory"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/nuclei/internal/config"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/nuclei/internal/usecase"
)

func init() {
	factory.Register("nuclei", func() factory.Source { return &Source{} })
}

// Source scrapes nuclei-templates from GitHub.
type Source struct{}

func (s *Source) Name() string { return "nuclei" }

func (s *Source) Policy() factory.FetchPolicy { return factory.PolicyPeriodic }

func (s *Source) Run(ctx context.Context, deps *factory.ScrapeDeps) error {
	cfg := config.FromEnv()
	pub, err := deps.Publisher("nuclei")
	if err != nil {
		return err
	}
	runner := usecase.NewRunner(deps.Log, pub, usecase.Options{
		MaxTemplates: cfg.MaxTemplates,
		YearsCSV:     cfg.YearsCSV,
	}, deps.Feeds, deps.Ledger)
	return runner.Run(ctx)
}
