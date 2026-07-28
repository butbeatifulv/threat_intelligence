package usecase

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/butbeautifulv/veil/discovery/pkg/githubraw"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/ledger"

	gh "github.com/butbeautifulv/veil/discovery/harvest/internal/sources/nuclei/internal/feeds/github"
)

func (r *Runner) githubListDir(ctx context.Context, owner, repo, path string) ([]gh.Content, error) {
	fc := r.feeds
	if fc == nil {
		fc = feeds.NewClient("", r.log)
	}
	return gh.NewClient(fc).ListDir(ctx, owner, repo, path)
}

func (r *Runner) fetchTemplateFile(ctx context.Context, owner, repo, ghPath string) ([]byte, bool, error) {
	cacheRel := filepath.Join("nuclei", strings.ReplaceAll(ghPath, "/", "__"))
	fc := r.feeds
	if fc == nil {
		fc = feeds.NewClient("", r.log)
	}
	var ledgerFn githubraw.FetchIfDueFunc
	if r.ledger != nil {
		ledgerFn = func(ctx context.Context, key, source, url, cachePath string, buildReq func() (*http.Request, error)) (githubraw.FetchResult, error) {
			res, err := feeds.FetchIfDue(ctx, fc, r.ledger, key, source, url, ledger.PolicyPeriodic, cachePath, buildReq)
			return githubraw.FetchResult{Body: res.Body, Skipped: res.Skipped, Unchanged: res.Unchanged}, err
		}
	}
	var lastErr error
	for _, ref := range feeds.GitHubRefs() {
		rawURL := feeds.GitHubRawURL(owner, repo, ref, ghPath)
		body, unchanged, err := githubraw.FetchFile(githubraw.FileParams{
			Ctx: ctx, Owner: owner, Repo: repo, Ref: ref, Path: ghPath,
			RawURL: rawURL, CacheRel: cacheRel, LedgerKey: "nuclei:file:" + ghPath, Source: "nuclei",
			UserAgent: "veil-scrape/1.0", Ledger: ledgerFn,
			FetchRaw: func(ctx context.Context, o, rep, rf, p string) ([]byte, error) {
				return feeds.GitHubFetchRaw(ctx, fc, o, rep, rf, p)
			},
			SkipErrFmt: "nuclei file %s skipped without cache",
		})
		if err == nil {
			return body, unchanged, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return nil, false, lastErr
	}
	return nil, false, fmt.Errorf("nuclei file %s: no ref succeeded", ghPath)
}
