// Package scrapesource registers the coderules scrape source with harvest/internal/factory.
package scrapesource

import (
	"context"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/factory"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/coderules/internal/config"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/coderules/internal/usecase"
)

func init() {
	factory.Register("coderules", func() factory.Source { return &Source{} })
}

// Source scrapes CWE, Semgrep, and CodeQL rule corpora.
type Source struct{}

func (s *Source) Name() string { return "coderules" }

func (s *Source) Policy() factory.FetchPolicy { return factory.PolicyPeriodic }

func (s *Source) Run(ctx context.Context, deps *factory.ScrapeDeps) error {
	cfg := config.FromEnv()
	pub, err := deps.Publisher("coderules")
	if err != nil {
		return err
	}
	runner := usecase.NewRunner(deps.Log, pub, usecase.Options{
		Sources:    cfg.Sources,
		MaxCWE:     cfg.MaxCWE,
		MaxSemgrep: cfg.MaxSemgrep,
		MaxCodeQL:  cfg.MaxCodeQL,
	}, deps.Feeds, deps.Ledger)
	return runner.Run(ctx)
}
