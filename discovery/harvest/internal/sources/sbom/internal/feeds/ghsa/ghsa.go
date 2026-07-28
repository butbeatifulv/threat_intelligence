package ghsa

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
)

const rawBase = "https://raw.githubusercontent.com/github/advisory-database/main"

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: 120 * time.Second}}
}

// NewClientFromEnv is kept for compatibility; GHSA uses open raw.githubusercontent.com only.
func NewClientFromEnv() *Client {
	return NewClient()
}

// CollectAdvisoryPaths discovers GHSA JSON paths via public git tree (no token).
func (c *Client) CollectAdvisoryPaths(ctx context.Context, fc *feeds.Client, maxPaths, minYear int) ([]string, error) {
	if maxPaths <= 0 {
		maxPaths = 50
	}
	if minYear <= 0 {
		minYear = 2017
	}
	if fc == nil {
		fc = feeds.NewClient("", nil)
	}
	var all []string
	var err error
	for _, ref := range []string{"main", "master"} {
		all, err = feeds.GitHubListTreePaths(ctx, fc, "github", "advisory-database", ref, "advisories/github-reviewed")
		if err == nil && len(all) > 0 {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	nowY := time.Now().UTC().Year()
	var out []string
	for _, p := range all {
		if len(out) >= maxPaths {
			break
		}
		if !strings.HasSuffix(p, ".json") {
			continue
		}
		if !strings.Contains(p, "/GHSA-") {
			continue
		}
		// advisories/github-reviewed/2024/01/GHSA-xxxx/GHSA-xxxx.json
		okYear := false
		for y := nowY; y >= minYear; y-- {
			if strings.Contains(p, fmt.Sprintf("/%d/", y)) {
				okYear = true
				break
			}
		}
		if !okYear {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

// FetchAdvisoryJSON downloads raw JSON (OSV schema) from raw.githubusercontent.com.
func (c *Client) FetchAdvisoryJSON(ctx context.Context, relPath string) (map[string]any, error) {
	relPath = strings.TrimPrefix(relPath, "/")
	u := rawBase + "/" + relPath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "veil-scrape/1.0")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ghsa raw: %s: %s", resp.Status, string(body[:min(200, len(body))]))
	}
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
