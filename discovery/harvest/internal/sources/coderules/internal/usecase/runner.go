package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"net/http"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/ledger"
	"github.com/butbeautifulv/veil/pkg/harvest"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/coderules/internal/feeds/cwe"
)

// Options for one run (CLI overrides env).
type Options struct {
	Sources            []string
	MaxCWE, MaxSemgrep int
	MaxCodeQL          int
}

// Runner runs coderules fetch → scrape.> (pipeline-worker → ingest.>).
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

func (r *Runner) enabled(name string) bool {
	for _, s := range r.opt.Sources {
		if s == name {
			return true
		}
	}
	return false
}

func (r *Runner) Run(ctx context.Context) error {
	if r.pub == nil {
		return fmt.Errorf("coderules: publisher required")
	}
	if r.enabled("cwe") {
		r.log.Info("ingesting CWE catalog (MITRE zip)…")
		bridge := &cweScrapeBridge{ctx: ctx, pub: r.pub}
		if r.feeds != nil {
			res, err := feeds.FetchIfDue(ctx, r.feeds, r.ledger, "coderules:cwe:mitre_zip", "coderules", cwe.ZipURL(), ledger.PolicyStatic, "coderules/cwe_mitre.zip", func() (*http.Request, error) {
				return http.NewRequestWithContext(ctx, http.MethodGet, cwe.ZipURL(), nil)
			})
			if err != nil {
				return err
			}
			if res.Unchanged {
				r.log.Info("CWE catalog unchanged, skip publish")
			} else {
				if res.Skipped && len(res.Body) == 0 {
					return fmt.Errorf("coderules:cwe skipped by ledger without cache")
				}
				if err := cwe.StreamMITREFromZip(ctx, bridge, r.opt.MaxCWE, res.Body); err != nil {
					return err
				}
			}
		} else if err := cwe.StreamMITRE(ctx, bridge, r.opt.MaxCWE); err != nil {
			return err
		}
	}
	if r.enabled("semgrep") {
		if err := r.runSemgrep(ctx); err != nil {
			return err
		}
	}
	if r.enabled("codeql") {
		if err := r.runCodeQL(ctx); err != nil {
			return err
		}
	}
	return nil
}

type cweScrapeBridge struct {
	ctx context.Context
	pub rawPublisher
}

func (b *cweScrapeBridge) UpsertCWECatalog(ctx context.Context, cweID, name, description, status string) error {
	pl := harvest.CoderulesCWERaw{ID: cweID, Name: name, Description: description, Status: status}
	return b.pub.Publish(b.ctx, harvest.KindCoderulesCWERaw, "coderules:cwe:"+cweID, pl)
}

func (r *Runner) runSemgrep(ctx context.Context) error {
	r.log.Info("ingesting Semgrep community rules (subset)…")
	const owner, repo = "semgrep", "semgrep-rules"
	seeds := []string{"python", "javascript", "java", "go", "csharp", "dockerfile", "yaml", "bash"}
	var q []string
	q = append(q, seeds...)
	n := 0
	for len(q) > 0 && n < r.opt.MaxSemgrep {
		dir := q[0]
		q = q[1:]
		items, err := r.githubListDir(ctx, owner, repo, dir)
		if err != nil {
			continue
		}
		for _, it := range items {
			if n >= r.opt.MaxSemgrep {
				break
			}
			if it.Type == "dir" && !strings.HasPrefix(it.Name, ".") {
				q = append(q, it.Path)
				continue
			}
			if it.Type != "file" || (!strings.HasSuffix(it.Name, ".yml") && !strings.HasSuffix(it.Name, ".yaml")) {
				continue
			}
			cacheRel := filepath.Join("semgrep", strings.ReplaceAll(it.Path, "/", "__"))
			raw, unchanged, err := r.fetchGitHubFile(ctx, owner, repo, it.Path, cacheRel)
			if err != nil || unchanged {
				continue
			}
			pl := harvest.CoderulesSemgrepRaw{Path: it.Path, RawYAML: string(raw)}
			if err := r.pub.Publish(ctx, harvest.KindCoderulesSemgrepRaw, "coderules:semgrep:"+it.Path, pl); err != nil {
				return err
			}
			n++
		}
	}
	r.log.Info("semgrep rules ingested", slog.Int("count", n))
	return nil
}

func (r *Runner) runCodeQL(ctx context.Context) error {
	r.log.Info("ingesting CodeQL queries (subset)…")
	const owner, repo = "github", "codeql"
	const path = "javascript/ql/src/Security/CWE-079"
	items, err := r.githubListDir(ctx, owner, repo, path)
	if err != nil {
		r.log.Warn("codeql list failed; skipping CodeQL subset", slog.String("path", path), slog.String("err", err.Error()))
		return nil
	}
	if len(items) == 0 {
		r.log.Warn("codeql list empty; skipping", slog.String("path", path))
		return nil
	}
	n := 0
	for _, it := range items {
		if n >= r.opt.MaxCodeQL {
			break
		}
		if it.Type != "file" || !strings.HasSuffix(it.Name, ".ql") {
			continue
		}
		cacheRel := filepath.Join("codeql", strings.ReplaceAll(it.Path, "/", "__"))
		raw, unchanged, err := r.fetchGitHubFile(ctx, owner, repo, it.Path, cacheRel)
		if err != nil || unchanged {
			continue
		}
		pl := harvest.CoderulesCodeQLRaw{Path: it.Path, Body: string(raw)}
		if err := r.pub.Publish(ctx, harvest.KindCoderulesCodeQLRaw, "coderules:codeql:"+it.Path, pl); err != nil {
			return err
		}
		n++
	}
	r.log.Info("codeql rules ingested", slog.Int("count", n))
	return nil
}
