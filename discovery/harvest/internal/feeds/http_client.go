package feeds

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/discovery/pkg/proxypool"
)

// ProxyOptions configures optional proxy pool from environment variables.
type ProxyOptions struct {
	URLsEnv string // e.g. "LOLA_PROXY_URLS"
	ModeEnv string // e.g. "LOLA_PROXY_MODE"
	Label   string // log label, e.g. "lola"
}

// SourceHTTPConfig builds a per-source feeds.Client with optional proxy pool.
type SourceHTTPConfig struct {
	CacheDir string
	Log      *slog.Logger
	Proxy    ProxyOptions
	Timeout  time.Duration
}

// NewSourceClient returns a feeds.Client with HTTP transport configured for one scrape source.
func NewSourceClient(cfg SourceHTTPConfig) *Client {
	if cfg.Log == nil {
		cfg.Log = slog.Default()
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 120 * time.Second
	}
	return &Client{
		HTTP:  NewHTTPClient(cfg.Log, cfg.Timeout, cfg.Proxy),
		Cache: cfg.CacheDir,
		Log:   cfg.Log,
	}
}

// NewHTTPClient builds an http.Client with TLS defaults and optional proxy pool.
func NewHTTPClient(log *slog.Logger, timeout time.Duration, proxy ProxyOptions) *http.Client {
	if log == nil {
		log = slog.Default()
	}
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSHandshakeTimeout = 30 * time.Second
	rt := http.RoundTripper(tr)
	if proxy.URLsEnv != "" {
		if env := strings.TrimSpace(os.Getenv(proxy.URLsEnv)); env != "" {
			p, err := proxypool.New(proxypool.SplitEnvList(env), 2*time.Minute)
			if err == nil {
				only := false
				if proxy.ModeEnv != "" {
					only = strings.EqualFold(strings.TrimSpace(os.Getenv(proxy.ModeEnv)), "only")
				}
				rt = proxypool.NewTransport(tr, p, only)
				label := proxy.Label
				if label == "" {
					label = "scrape"
				}
				log.Info(label+" proxy pool enabled", slog.Int("count", len(proxypool.SplitEnvList(env))))
			} else {
				label := proxy.Label
				if label == "" {
					label = "scrape"
				}
				log.Warn(label+" proxy pool invalid; running direct", slog.String("err", err.Error()))
			}
		}
	}
	return &http.Client{Timeout: timeout, Transport: rt}
}

// FetchWithRetry GETs a URL with exponential backoff; uses Client cache when set.
func (c *Client) FetchWithRetry(ctx context.Context, downloadURL, cachePath, userAgent string, maxAttempts int) ([]byte, error) {
	if c == nil || c.HTTP == nil {
		return nil, fmt.Errorf("feeds: client required")
	}
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	if b, ok := c.ReadCache(cachePath); ok {
		return b, nil
	}
	backoff := 1 * time.Second
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
		if err != nil {
			return nil, err
		}
		if userAgent != "" {
			req.Header.Set("User-Agent", userAgent)
		}
		resp, err := c.HTTP.Do(req)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil, err
			}
			lastErr = err
			if !sleepOrCancel(ctx, backoff) {
				return nil, ctx.Err()
			}
			backoff *= 2
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
			_ = resp.Body.Close()
			return nil, fmt.Errorf("download %s: %d %s", downloadURL, resp.StatusCode, string(b))
		}
		b, rerr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if rerr != nil {
			lastErr = rerr
			if !sleepOrCancel(ctx, backoff) {
				return nil, ctx.Err()
			}
			backoff *= 2
			continue
		}
		_ = c.WriteCache(cachePath, b)
		return b, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("download failed: %s", downloadURL)
}

// SleepOrCancel waits for d or until ctx is canceled.
func SleepOrCancel(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return true
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func sleepOrCancel(ctx context.Context, d time.Duration) bool {
	return SleepOrCancel(ctx, d)
}
