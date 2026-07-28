package feeds

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Client is a shared HTTP fetcher with optional disk cache (L1).
type Client struct {
	HTTP  *http.Client
	Cache string
	Log   *slog.Logger
}

func NewClient(cacheDir string, log *slog.Logger) *Client {
	if log == nil {
		log = slog.Default()
	}
	return &Client{
		HTTP:  &http.Client{Timeout: 120 * time.Second},
		Cache: cacheDir,
		Log:   log,
	}
}

// ReadCache returns cached bytes when cachePath is set and file exists.
func (c *Client) ReadCache(cachePath string) ([]byte, bool) {
	if c.Cache == "" || cachePath == "" {
		return nil, false
	}
	fn := filepath.Join(c.Cache, filepath.FromSlash(cachePath))
	b, err := os.ReadFile(fn) // #nosec G304 -- cachePath is a static literal or fmt-built from int params at every call site (see feeds.FetchIfDue callers), never attacker/network-controlled
	if err != nil || len(b) == 0 {
		return nil, false
	}
	return b, true
}

// WriteCache stores body at cachePath under Cache root.
func (c *Client) WriteCache(cachePath string, body []byte) error {
	if c.Cache == "" || cachePath == "" || len(body) == 0 {
		return nil
	}
	fn := filepath.Join(c.Cache, filepath.FromSlash(cachePath))
	if err := os.MkdirAll(filepath.Dir(fn), 0o750); err != nil {
		return err
	}
	return os.WriteFile(fn, body, 0o600)
}

// DoGET performs GET with optional cache read/write.
func (c *Client) DoGET(req *http.Request, cachePath string) ([]byte, error) {
	if b, ok := c.ReadCache(cachePath); ok {
		return b, nil
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, &HTTPStatusError{Code: resp.StatusCode, Body: string(b)}
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	_ = c.WriteCache(cachePath, b)
	return b, nil
}

type HTTPStatusError struct {
	Code int
	Body string
}

func (e *HTTPStatusError) Error() string {
	return http.StatusText(e.Code)
}
