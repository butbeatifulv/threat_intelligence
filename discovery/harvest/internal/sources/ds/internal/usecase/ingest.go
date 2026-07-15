package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/ledger"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/ds/internal/repository"
)

type Ingestor struct {
	store  repository.GraphStore
	logger *slog.Logger
	feeds  *feeds.Client
	ledger *ledger.Store
}

func NewIngestor(store repository.GraphStore, logger *slog.Logger, fc *feeds.Client, led *ledger.Store) *Ingestor {
	return &Ingestor{store: store, logger: logger, feeds: fc, ledger: led}
}

func dsGitRef() string {
	return firstNonEmpty(os.Getenv("DS_GITHUB_REF"), "master")
}

func (u *Ingestor) Run(ctx context.Context) error {
	if err := u.store.EnsureSchema(ctx); err != nil {
		return err
	}
	maxSigma := envInt("DS_MAX_SIGMA", 250)
	maxYara := envInt("DS_MAX_YARA", 120)
	maxAtomic := envInt("DS_MAX_ATOMIC", 200)

	if err := u.ingestSigmaDir(ctx, "SigmaHQ", "sigma", "rules/windows/process_creation", maxSigma); err != nil {
		return err
	}
	if err := u.ingestYaraDir(ctx, "Neo23x0", "signature-base", firstNonEmpty(os.Getenv("DS_YARA_PATH"), "yara"), maxYara); err != nil {
		u.logger.Warn("yara ingest primary path failed", slog.String("err", err.Error()))
		if err2 := u.ingestYaraDir(ctx, "Neo23x0", "signature-base", "iocs/yara", maxYara); err2 != nil {
			return err2
		}
	}
	if err := u.ingestAtomicTests(ctx, maxAtomic); err != nil {
		return err
	}
	if err := u.ingestCalderaAbilities(ctx); err != nil {
		return err
	}
	return nil
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func (u *Ingestor) ingestSigmaDir(ctx context.Context, owner, repo, path string, limit int) error {
	u.logger.Info("ingesting Sigma rules", slog.String("path", path), slog.Int("limit", limit))
	items, err := feeds.GitHubListDir(ctx, u.feeds, owner, repo, path)
	if err != nil {
		return err
	}
	n := 0
	for _, it := range items {
		if n >= limit {
			break
		}
		if it.Type != "file" || (!strings.HasSuffix(it.Name, ".yml") && !strings.HasSuffix(it.Name, ".yaml")) {
			continue
		}
		raw, unchanged, err := u.fetchBytes(ctx, owner, repo, it.Path, filepath.Join("sigma", strings.ReplaceAll(it.Path, "/", "__")))
		if err != nil || unchanged {
			continue
		}
		if err := u.store.UpsertSigmaRaw(ctx, it.Path, string(raw)); err != nil {
			return err
		}
		n++
	}
	u.logger.Info("sigma ingest done", slog.Int("count", n))
	return nil
}

func tagsToJSON(v any) string {
	arr, ok := v.([]any)
	if !ok {
		return "[]"
	}
	ss := make([]string, 0, len(arr))
	for _, x := range arr {
		if s, ok := x.(string); ok {
			ss = append(ss, s)
		}
	}
	b, _ := json.Marshal(ss)
	return string(b)
}

func (u *Ingestor) ingestYaraDir(ctx context.Context, owner, repo, path string, limit int) error {
	u.logger.Info("ingesting YARA rules", slog.Int("limit", limit))
	items, err := feeds.GitHubListDir(ctx, u.feeds, owner, repo, path)
	if err != nil {
		return err
	}
	n := 0
	for _, it := range items {
		if n >= limit {
			break
		}
		if it.Type != "file" || !strings.HasSuffix(it.Name, ".yar") {
			continue
		}
		raw, unchanged, err := u.fetchBytes(ctx, owner, repo, it.Path, filepath.Join("yara", strings.ReplaceAll(it.Path, "/", "__")))
		if err != nil || unchanged {
			continue
		}
		body := string(raw)
		name := parseYaraRuleName(body, it.Name)
		if err := u.store.UpsertYaraRaw(ctx, it.Path, name, body); err != nil {
			return err
		}
		n++
	}
	u.logger.Info("yara ingest done", slog.Int("count", n))
	return nil
}

func parseYaraRuleName(body, fallback string) string {
	lines := strings.Split(body, "\n")
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if strings.HasPrefix(ln, "rule ") {
			p := strings.TrimPrefix(ln, "rule ")
			if idx := strings.IndexAny(p, " \t{"); idx > 0 {
				return strings.TrimSpace(p[:idx])
			}
			return strings.TrimSpace(p)
		}
	}
	return strings.TrimSuffix(fallback, ".yar")
}

func (u *Ingestor) ingestAtomicTests(ctx context.Context, limit int) error {
	if limit <= 0 {
		u.logger.Info("atomic red team skipped (DS_MAX_ATOMIC<=0)")
		return nil
	}
	u.logger.Info("ingesting Atomic Red Team YAML", slog.Int("limit", limit))
	items, err := feeds.GitHubListDir(ctx, u.feeds, "redcanaryco", "atomic-red-team", "atomics")
	if err != nil {
		u.logger.Warn("atomic red team listing skipped", slog.String("err", err.Error()))
		return nil
	}
	n := 0
	for _, dir := range items {
		if n >= limit {
			break
		}
		if dir.Type != "dir" {
			continue
		}
		// technique folder contains Txxxx.yaml
		sub, err := feeds.GitHubListDir(ctx, u.feeds, "redcanaryco", "atomic-red-team", dir.Path)
		if err != nil {
			continue
		}
		for _, f := range sub {
			if n >= limit {
				break
			}
			if f.Type != "file" || (!strings.HasSuffix(f.Name, ".yaml") && !strings.HasSuffix(f.Name, ".yml")) {
				continue
			}
			raw, unchanged, err := u.fetchBytes(ctx, "redcanaryco", "atomic-red-team", f.Path, filepath.Join("atomic", strings.ReplaceAll(f.Path, "/", "__")))
			if err != nil || unchanged {
				continue
			}
			if err := u.store.UpsertAtomicRaw(ctx, f.Path, string(raw)); err != nil {
				return err
			}
			n++
			if n >= limit {
				break
			}
		}
		if n >= limit {
			break
		}
	}
	u.logger.Info("atomic ingest done", slog.Int("count", n))
	return nil
}

func (u *Ingestor) fetchBytes(ctx context.Context, owner, repo, ghPath, cacheRel string) (body []byte, unchanged bool, err error) {
	ref := dsGitRef()
	rawURL := feeds.GitHubRawURL(owner, repo, ref, ghPath)
	key := fmt.Sprintf("gh:file:%s:%s:%s", owner, repo, ghPath)
	res, err := feeds.FetchIfDue(ctx, u.feeds, u.ledger, key, "ds", rawURL, ledger.PolicyPeriodic, cacheRel, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "veil-scrape/1.0")
		return req, nil
	})
	if err != nil {
		return nil, false, err
	}
	if res.Unchanged {
		return nil, true, nil
	}
	if res.Skipped && len(res.Body) == 0 {
		return nil, false, fmt.Errorf("download %s skipped without cache", rawURL)
	}
	return res.Body, false, nil
}
