package feeds

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/ledger"
)

// FetchResult is the outcome of FetchIfDue.
type FetchResult struct {
	Body      []byte
	SHA256    string
	Skipped   bool // ledger said not due (no HTTP)
	Unchanged bool // content hash matched ledger (skip publish)
}

// FetchIfDue checks Vitess ledger (when set), fetches URL, records metadata.
func FetchIfDue(
	ctx context.Context,
	c *Client,
	led CrawlLedger,
	resourceKey, source, url string,
	policy ledger.FetchPolicy,
	cachePath string,
	buildReq func() (*http.Request, error),
) (FetchResult, error) {
	force := strings.TrimSpace(os.Getenv("SCRAPE_FORCE_REFETCH")) == "1"
	minRefetch := parseMinRefetch()

	var prevSHA string
	if led != nil {
		var err error
		prevSHA, err = led.GetContentSHA(ctx, resourceKey)
		if err != nil {
			return FetchResult{}, err
		}
		ok, err := led.ShouldFetch(ctx, resourceKey, policy, minRefetch, force)
		if err != nil {
			return FetchResult{}, err
		}
		if !ok {
			if b, hit := c.ReadCache(cachePath); hit {
				sum := sha256.Sum256(b)
				sha := hex.EncodeToString(sum[:])
				unchanged := prevSHA != "" && prevSHA == sha
				return FetchResult{Body: b, SHA256: sha, Skipped: true, Unchanged: unchanged}, nil
			}
			if c.Log != nil {
				c.Log.Info("ledger skip but cache miss; refetching",
					slog.String("resource_key", resourceKey),
					slog.String("cache_path", cachePath),
				)
			}
			// Fall through to HTTP fetch — ledger without blob is invalid.
		}
	}

	req, err := buildReq()
	if err != nil {
		return FetchResult{}, err
	}
	body, err := c.DoGET(req, cachePath)
	if err != nil {
		return FetchResult{}, err
	}
	sum := sha256.Sum256(body)
	sha := hex.EncodeToString(sum[:])
	unchanged := prevSHA != "" && prevSHA == sha

	if led != nil {
		if err := led.RecordFetch(ctx, resourceKey, source, url, policy, sha); err != nil {
			return FetchResult{}, err
		}
	}
	return FetchResult{Body: body, SHA256: sha, Unchanged: unchanged}, nil
}

func parseMinRefetch() time.Duration {
	v := strings.TrimSpace(os.Getenv("SCRAPE_MIN_REFETCH_AFTER"))
	if v == "" {
		v = "24h"
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		if h, e := strconv.Atoi(v); e == nil && h > 0 {
			return time.Duration(h) * time.Hour
		}
		return 24 * time.Hour
	}
	return d
}

// OpenLedgerFromEnv opens ledger when VITESS_DSN or MYSQL_DSN is set.
func OpenLedgerFromEnv(ctx context.Context) (*ledger.Store, error) {
	if strings.TrimSpace(os.Getenv("VITESS_DSN")) == "" && strings.TrimSpace(os.Getenv("MYSQL_DSN")) == "" {
		return nil, nil
	}
	st, err := ledger.OpenFromEnv()
	if err != nil {
		return nil, err
	}
	if err := st.EnsureSchema(ctx); err != nil {
		_ = st.Close()
		return nil, err
	}
	return st, nil
}
