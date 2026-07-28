// Package cvesource lists CVE ids for SBOM OSV ingest without Neo4j (NATS mode).
package cvesource

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Lister returns CVE identifiers (CVE-YYYY-NNNN+) for the OSV fetch loop.
type Lister interface {
	ListCVEs(ctx context.Context, limit int) ([]string, error)
}

// FromFile returns CVE ids from a local file (one per line; # starts a comment).
type FromFile struct {
	Path string
}

func (f FromFile) ListCVEs(ctx context.Context, limit int) ([]string, error) {
	if strings.TrimSpace(f.Path) == "" {
		return nil, fmt.Errorf("cvesource: empty file path")
	}
	data, err := os.ReadFile(f.Path)
	if err != nil {
		return nil, fmt.Errorf("cvesource: read %q: %w", f.Path, err)
	}
	return parseCVEList(ctx, string(data), limit)
}

// FromHTTP fetches a text/plain or similar body (one CVE per line).
type FromHTTP struct {
	URL    string
	Client *http.Client
}

func (h FromHTTP) ListCVEs(ctx context.Context, limit int) ([]string, error) {
	u := strings.TrimSpace(h.URL)
	if u == "" {
		return nil, fmt.Errorf("cvesource: empty URL")
	}
	cli := h.Client
	if cli == nil {
		cli = &http.Client{Timeout: 120 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cvesource: get %q: %w", u, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slurp, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("cvesource: %s: %s", resp.Status, strings.TrimSpace(string(slurp)))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	return parseCVEList(ctx, string(body), limit)
}

func parseCVEList(ctx context.Context, raw string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 100
	}
	var out []string
	sc := bufio.NewScanner(strings.NewReader(raw))
	for sc.Scan() {
		select {
		case <-ctx.Done():
			return out, ctx.Err()
		default:
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(strings.ToUpper(line), "CVE-") {
			continue
		}
		out = append(out, strings.ToUpper(line))
		if len(out) >= limit {
			break
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("cvesource: no CVE lines found")
	}
	return out, nil
}
