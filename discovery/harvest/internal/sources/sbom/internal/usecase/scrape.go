package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/ledger"
	"github.com/butbeautifulv/veil/pkg/harvest"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/sbom/internal/cvesource"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/sbom/internal/feeds/ghsa"
)

// Options overrides env defaults for one run (from CLI flags).
type Options struct {
	Sources     []string
	MaxCVE      int
	MaxGHSA     int
	GHSAMinYear int
}

// Runner orchestrates SBOM fetch → scrape.> (pipeline-worker → ingest.> → Neo4j).
type Runner struct {
	log    *slog.Logger
	cves   cvesource.Lister
	pub    rawPublisher
	opt    Options
	feeds  *feeds.Client
	ledger *ledger.Store
}

func NewRunner(log *slog.Logger, cves cvesource.Lister, pub rawPublisher, opt Options, fc *feeds.Client, led *ledger.Store) *Runner {
	return &Runner{log: log, cves: cves, pub: pub, opt: opt, feeds: fc, ledger: led}
}

func (r *Runner) sourceEnabled(name string) bool {
	for _, s := range r.opt.Sources {
		if s == name {
			return true
		}
	}
	return false
}

func (r *Runner) Run(ctx context.Context) error {
	if r.pub == nil {
		return fmt.Errorf("usecase: publisher required")
	}
	if r.cves == nil {
		return fmt.Errorf("usecase: set SBOM_CVE_LIST_FILE or SBOM_CVE_LIST_URL for CVE list")
	}
	if r.sourceEnabled("osv") {
		if err := r.runOSV(ctx); err != nil {
			return err
		}
	}
	if r.sourceEnabled("ghsa") {
		if err := r.runGHSA(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) runOSV(ctx context.Context) error {
	cves, err := r.cves.ListCVEs(ctx, r.opt.MaxCVE)
	if err != nil {
		return err
	}
	for i, cve := range cves {
		osvURL := "https://api.osv.dev/v1/vulns/" + cve
		key := "osv:" + cve
		cacheRel := "osv/" + cve + ".json"

		var doc map[string]any
		if r.feeds != nil {
			res, err := feeds.FetchIfDue(ctx, r.feeds, r.ledger, key, "sbom", osvURL, ledger.PolicyPeriodic, cacheRel, func() (*http.Request, error) {
				return http.NewRequestWithContext(ctx, http.MethodGet, osvURL, nil)
			})
			if err != nil {
				r.log.Warn("osv fetch", slog.String("cve", cve), slog.String("err", err.Error()))
				continue
			}
			if res.Unchanged {
				continue
			}
			if res.Skipped && len(res.Body) == 0 {
				r.log.Warn("osv skipped without cache", slog.String("cve", cve))
				continue
			}
			if err := json.Unmarshal(res.Body, &doc); err != nil {
				r.log.Warn("osv decode", slog.String("cve", cve), slog.String("err", err.Error()))
				continue
			}
		} else {
			r.log.Warn("osv: feeds client required for ledger path", slog.String("cve", cve))
			continue
		}

		id, _ := doc["id"].(string)
		raw, _ := json.Marshal(doc)
		pl := harvest.SBOMOSVRaw{CVE: cve, OSVID: id, RawJSON: string(raw)}
		if err := r.pub.Publish(ctx, harvest.KindSBOMOSVJSON, "sbom:osv:"+cve, pl); err != nil {
			return fmt.Errorf("publish osv %s: %w", cve, err)
		}
		if (i+1)%20 == 0 {
			r.log.Info("osv progress", slog.Int("done", i+1), slog.Int("total", len(cves)))
		}
		if !feeds.SleepOrCancel(ctx, 150*time.Millisecond) {
			return ctx.Err()
		}
	}
	return nil
}

func (r *Runner) runGHSA(ctx context.Context) error {
	gc := ghsa.NewClientFromEnv()
	paths, err := gc.CollectAdvisoryPaths(ctx, r.feeds, r.opt.MaxGHSA, r.opt.GHSAMinYear)
	if err != nil {
		return err
	}
	for i, p := range paths {
		doc, err := gc.FetchAdvisoryJSON(ctx, p)
		if err != nil {
			r.log.Warn("ghsa fetch", slog.String("path", p), slog.String("err", err.Error()))
			continue
		}
		pl := harvest.SBOMGHSARaw{Path: p, Doc: doc}
		if err := r.pub.Publish(ctx, harvest.KindSBOMGHSAPath, "sbom:ghsa:"+p, pl); err != nil {
			return fmt.Errorf("publish ghsa %s: %w", p, err)
		}
		if (i+1)%10 == 0 {
			r.log.Info("ghsa progress", slog.Int("done", i+1), slog.Int("total", len(paths)))
		}
		if !feeds.SleepOrCancel(ctx, 100*time.Millisecond) {
			return ctx.Err()
		}
	}
	return nil
}
