// Package scrapesource registers the sbom scrape source with harvest/internal/factory.
package scrapesource

import (
	"context"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/factory"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/sbom/internal/config"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/sbom/internal/cvesource"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/sbom/internal/usecase"
)

func init() {
	factory.Register("sbom", func() factory.Source { return &Source{} })
}

// Source scrapes OSV and GHSA advisories.
type Source struct{}

func (s *Source) Name() string { return "sbom" }

func (s *Source) Policy() factory.FetchPolicy { return factory.PolicyPeriodic }

func (s *Source) Run(ctx context.Context, deps *factory.ScrapeDeps) error {
	cfg := config.FromEnv()
	cveSrc, err := cvesource.New(cfg.CVEListFile, cfg.CVEListURL)
	if err != nil {
		return err
	}
	pub, err := deps.Publisher("sbom")
	if err != nil {
		return err
	}
	runner := usecase.NewRunner(deps.Log, cveSrc, pub, usecase.Options{
		Sources:     cfg.Sources,
		MaxCVE:      cfg.MaxCVE,
		MaxGHSA:     cfg.MaxGHSA,
		GHSAMinYear: cfg.GHSAMinYear,
	}, deps.Feeds, deps.Ledger)
	return runner.Run(ctx)
}
