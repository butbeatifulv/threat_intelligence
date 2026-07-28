// Package githubraw fetches GitHub raw files with optional ledger-backed dedup (scrape only).
package githubraw

import (
	"context"
	"fmt"
	"net/http"
)

// FetchResult mirrors scrape/feeds.FetchResult for ledger-backed fetches.
type FetchResult struct {
	Body      []byte
	Skipped   bool
	Unchanged bool
}

// FetchIfDueFunc performs a ledger-aware HTTP fetch (implemented by scrape/feeds).
type FetchIfDueFunc func(
	ctx context.Context,
	resourceKey, source, url, cachePath string,
	buildReq func() (*http.Request, error),
) (FetchResult, error)

// FetchRawFunc fetches a GitHub raw file without ledger (implemented by scrape/feeds).
type FetchRawFunc func(ctx context.Context, owner, repo, ref, path string) ([]byte, error)

// FileParams configures a single GitHub raw file fetch with optional ledger dedup.
type FileParams struct {
	Ctx        context.Context
	Owner      string
	Repo       string
	Ref        string
	Path       string
	RawURL     string
	CacheRel   string
	LedgerKey  string
	Source     string
	UserAgent  string
	Ledger     FetchIfDueFunc
	FetchRaw   FetchRawFunc
	SkipErrFmt string // e.g. "nuclei file %s skipped without cache"
}

// FetchFile returns body, unchanged (skip publish), and error.
func FetchFile(p FileParams) ([]byte, bool, error) {
	if p.Ledger == nil {
		raw, err := p.FetchRaw(p.Ctx, p.Owner, p.Repo, p.Ref, p.Path)
		return raw, false, err
	}
	res, err := p.Ledger(p.Ctx, p.LedgerKey, p.Source, p.RawURL, p.CacheRel, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(p.Ctx, http.MethodGet, p.RawURL, nil)
		if err != nil {
			return nil, err
		}
		if p.UserAgent != "" {
			req.Header.Set("User-Agent", p.UserAgent)
		}
		return req, nil
	})
	if err != nil {
		return nil, false, err
	}
	if res.Unchanged {
		return nil, true, nil
	}
	if res.Skipped && len(res.Body) == 0 {
		msg := p.SkipErrFmt
		if msg == "" {
			msg = "github file %s skipped without cache"
		}
		return nil, false, fmt.Errorf(msg, p.Path)
	}
	return res.Body, false, nil
}
