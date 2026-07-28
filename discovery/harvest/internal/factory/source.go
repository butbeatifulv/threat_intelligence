// Package factory registers scrape sources and shared ScrapeDeps.
package factory

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/ledger"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/scrapepub"
)

// FetchPolicy aliases ledger policies for sources.
type FetchPolicy = ledger.FetchPolicy

const (
	PolicyStatic   = ledger.PolicyStatic
	PolicyPeriodic = ledger.PolicyPeriodic
	PolicyDaily    = ledger.PolicyDaily
)

// RawPublisher publishes harvest for one domain source and subject.
type RawPublisher = scrapepub.RawPublisher

// Source is one scrape feed run (vuln NVD, ds sigma, ti kev, …).
type Source interface {
	Name() string
	Policy() FetchPolicy
	Run(ctx context.Context, deps *ScrapeDeps) error
}

// ScrapeDeps is shared infrastructure for all scrapers.
type ScrapeDeps struct {
	Ledger     *ledger.Store
	Feeds      *feeds.Client
	Log        *slog.Logger
	Publishers map[string]RawPublisher
}

// Publisher returns the raw publisher for a source name.
func (d *ScrapeDeps) Publisher(name string) (RawPublisher, error) {
	if d == nil || d.Publishers == nil {
		return nil, fmt.Errorf("scrape deps: no publishers")
	}
	p, ok := d.Publishers[name]
	if !ok {
		return nil, fmt.Errorf("scrape deps: missing publisher for %q", name)
	}
	return p, nil
}

// Registry runs registered sources in order.
type Registry struct {
	sources []Source
}

func NewRegistry(sources ...Source) *Registry {
	return &Registry{sources: sources}
}

func (r *Registry) RunAll(ctx context.Context, deps *ScrapeDeps) error {
	failFast := strings.EqualFold(strings.TrimSpace(os.Getenv("SCRAPE_FAIL_FAST")), "true") ||
		strings.TrimSpace(os.Getenv("SCRAPE_FAIL_FAST")) == "1"
	var failed []string
	for _, s := range r.sources {
		deps.Log.Info("scrape source start", slog.String("source", s.Name()), slog.String("policy", string(s.Policy())))
		if err := s.Run(ctx, deps); err != nil {
			if failFast {
				return fmt.Errorf("%s: %w", s.Name(), err)
			}
			deps.Log.Error("scrape source failed; continuing", slog.String("source", s.Name()), slog.String("error", err.Error()))
			failed = append(failed, s.Name())
			continue
		}
		deps.Log.Info("scrape source done", slog.String("source", s.Name()))
	}
	if len(failed) > 0 {
		return fmt.Errorf("scrape: %d source(s) failed: %s", len(failed), strings.Join(failed, ", "))
	}
	return nil
}
