package feeds

import (
	"bytes"
	"context"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	scrapecache "github.com/butbeautifulv/veil/discovery/harvest/internal/cache"
	scrapefeeds "github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/ledger"
	"github.com/butbeautifulv/veil/discovery/pkg/proxypool"
)

const tiUserAgent = "veil-ti/1.0"

// Fetcher performs ledger-aware HTTP fetches for TI feeds.
type Fetcher struct {
	Feeds  *scrapefeeds.Client
	Ledger scrapefeeds.CrawlLedger
	Delay  time.Duration
}

// NewFetcher builds HTTP transport (optional proxy pool) and wires the shared feeds client.
func NewFetcher(logger *slog.Logger, fc *scrapefeeds.Client, led scrapefeeds.CrawlLedger) *Fetcher {
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.TLSHandshakeTimeout = 30 * time.Second
	var rt http.RoundTripper = base
	if env := strings.TrimSpace(os.Getenv("TI_PROXY_URLS")); env != "" {
		p, err := proxypool.New(proxypool.SplitEnvList(env), 2*time.Minute)
		if err == nil {
			only := strings.EqualFold(strings.TrimSpace(os.Getenv("TI_PROXY_MODE")), "only")
			rt = proxypool.NewTransport(base, p, only)
			logger.Info("ti proxy pool enabled", slog.Int("count", len(proxypool.SplitEnvList(env))))
		} else {
			logger.Warn("ti proxy pool invalid; running without proxy", slog.String("err", err.Error()))
		}
	}
	cache := firstNonEmpty(os.Getenv("TI_CACHE_DIR"), scrapecache.DefaultDir())
	if fc == nil {
		fc = scrapefeeds.NewClient(cache, logger)
	}
	if fc.Cache == "" {
		fc.Cache = cache
	}
	hc := &http.Client{Timeout: 120 * time.Second, Transport: rt}
	fc.HTTP = hc
	return &Fetcher{
		Feeds:  fc,
		Ledger: led,
		Delay:  parseDelayEnv(os.Getenv("TI_REQUEST_DELAY"), 1200*time.Millisecond),
	}
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func parseDelayEnv(v string, def time.Duration) time.Duration {
	v = strings.TrimSpace(v)
	if v == "" {
		return def
	}
	if d, err := time.ParseDuration(v); err == nil && d >= 0 {
		return d
	}
	if ms, err := strconv.Atoi(v); err == nil && ms >= 0 {
		return time.Duration(ms) * time.Millisecond
	}
	return def
}

func (f *Fetcher) preNetworkDelay(ctx context.Context) error {
	if f.Delay <= 0 {
		return nil
	}
	extra := time.Duration(0)
	if f.Delay > 4 {
		extra = time.Duration(rand.Int64N(int64(f.Delay / 4))) // #nosec G404 -- request pacing jitter, not security-sensitive
	}
	if !scrapefeeds.SleepOrCancel(ctx, f.Delay+extra) {
		return ctx.Err()
	}
	return nil
}

// FetchLedger wraps FetchIfDue for GET feeds.
func (f *Fetcher) FetchLedger(ctx context.Context, key, url, cacheRel string, policy ledger.FetchPolicy) (scrapefeeds.FetchResult, error) {
	if err := f.preNetworkDelay(ctx); err != nil {
		return scrapefeeds.FetchResult{}, err
	}
	return scrapefeeds.FetchIfDue(ctx, f.Feeds, f.Ledger, key, "ti", url, policy, cacheRel, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", tiUserAgent)
		return req, nil
	})
}

// FetchLedgerPOST wraps FetchIfDue for JSON POST feeds (ThreatFox API, MalwareBazaar).
func (f *Fetcher) FetchLedgerPOST(ctx context.Context, key, url, cacheRel string, policy ledger.FetchPolicy, authKey string, jsonBody []byte) (scrapefeeds.FetchResult, error) {
	if err := f.preNetworkDelay(ctx); err != nil {
		return scrapefeeds.FetchResult{}, err
	}
	body := append([]byte(nil), jsonBody...)
	return scrapefeeds.FetchIfDue(ctx, f.Feeds, f.Ledger, key, "ti", url, policy, cacheRel, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", tiUserAgent)
		req.Header.Set("Content-Type", "application/json")
		if strings.TrimSpace(authKey) != "" {
			req.Header.Set("Auth-Key", authKey)
		}
		return req, nil
	})
}
