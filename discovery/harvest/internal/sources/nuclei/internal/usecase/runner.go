package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/ledger"
	"github.com/butbeautifulv/veil/pkg/harvest"
)

// Options for one run (CLI overrides env).
type Options struct {
	MaxTemplates int
	YearsCSV     string
}

// Runner runs nuclei template fetch → scrape.>.
type Runner struct {
	log    *slog.Logger
	pub    rawPublisher
	opt    Options
	feeds  *feeds.Client
	ledger *ledger.Store
}

func NewRunner(log *slog.Logger, pub rawPublisher, opt Options, fc *feeds.Client, led *ledger.Store) *Runner {
	return &Runner{log: log, pub: pub, opt: opt, feeds: fc, ledger: led}
}

func (r *Runner) Run(ctx context.Context) error {
	if r.pub == nil {
		return fmt.Errorf("nuclei: publisher required")
	}

	const owner, repo = "projectdiscovery", "nuclei-templates"
	n := 0
	for _, y := range strings.Split(r.opt.YearsCSV, ",") {
		if n >= r.opt.MaxTemplates {
			break
		}
		year := strings.TrimSpace(y)
		if year == "" {
			continue
		}
		base := "http/cves/" + year
		items, err := r.githubListDir(ctx, owner, repo, base)
		if err != nil {
			r.log.Warn("list dir", slog.String("base", base), slog.String("err", err.Error()))
			continue
		}
		if len(items) == 0 {
			r.log.Warn("nuclei list empty", slog.String("base", base))
			continue
		}
		for _, it := range items {
			if n >= r.opt.MaxTemplates {
				break
			}
			if it.Type != "file" || (!strings.HasSuffix(it.Name, ".yaml") && !strings.HasSuffix(it.Name, ".yml")) {
				continue
			}
			raw, unchanged, err := r.fetchTemplateFile(ctx, owner, repo, it.Path)
			if err != nil || unchanged {
				continue
			}
			pl := harvest.NucleiTemplateRaw{Path: it.Path, RawYAML: string(raw)}
			if err := r.pub.Publish(ctx, harvest.KindNucleiTemplateRaw, "nuclei:"+it.Path, pl); err != nil {
				return err
			}
			n++
		}
	}
	if n == 0 {
		r.log.Warn("nuclei templates ingested", slog.Int("count", 0), slog.String("years", r.opt.YearsCSV))
	} else {
		r.log.Info("nuclei templates ingested", slog.Int("count", n))
	}
	return nil
}
