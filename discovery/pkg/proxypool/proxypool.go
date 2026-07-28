// Package proxypool rotates HTTP proxies for harvest sources (scrape only).
package proxypool

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Pool struct {
	mu       sync.Mutex
	proxies  []*url.URL
	next     int
	cooldown time.Duration
	state    map[string]time.Time
}

func New(proxyURLs []string, cooldown time.Duration) (*Pool, error) {
	var parsed []*url.URL
	for _, s := range proxyURLs {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		u, err := url.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("parse proxy url %q: %w", s, err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return nil, fmt.Errorf("unsupported proxy scheme %q in %q (use http/https)", u.Scheme, s)
		}
		if u.Host == "" {
			return nil, fmt.Errorf("proxy url %q missing host", s)
		}
		parsed = append(parsed, u)
	}
	if len(parsed) == 0 {
		return nil, errors.New("empty proxy list")
	}
	if cooldown <= 0 {
		cooldown = 2 * time.Minute
	}
	return &Pool{
		proxies:  parsed,
		cooldown: cooldown,
		state:    make(map[string]time.Time, len(parsed)),
	}, nil
}

func SplitEnvList(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	v = strings.ReplaceAll(v, "\n", ",")
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (p *Pool) pickLocked(now time.Time) *url.URL {
	n := len(p.proxies)
	for i := 0; i < n; i++ {
		idx := (p.next + i) % n
		u := p.proxies[idx]
		key := u.String()
		if ua, ok := p.state[key]; ok && ua.After(now) {
			continue
		}
		p.next = (idx + 1) % n
		return u
	}
	u := p.proxies[p.next%n]
	p.next = (p.next + 1) % n
	return u
}

func (p *Pool) nextUsableAfterLocked(now time.Time) time.Time {
	var min time.Time
	for _, u := range p.proxies {
		if ua, ok := p.state[u.String()]; ok && ua.After(now) {
			if min.IsZero() || ua.Before(min) {
				min = ua
			}
		} else {
			return time.Time{}
		}
	}
	return min
}

func (p *Pool) MarkBad(proxy *url.URL) {
	if proxy == nil {
		return
	}
	now := time.Now()
	jitter := time.Duration(rand.IntN(750)+250) * time.Millisecond // #nosec G404 -- cooldown jitter timing, not security-sensitive
	p.mu.Lock()
	p.state[proxy.String()] = now.Add(p.cooldown + jitter)
	p.mu.Unlock()
}

func (p *Pool) MarkMaybeBad(proxy *url.URL) {
	if proxy == nil {
		return
	}
	now := time.Now()
	jitter := time.Duration(rand.IntN(400)+150) * time.Millisecond // #nosec G404 -- cooldown jitter timing, not security-sensitive
	p.mu.Lock()
	p.state[proxy.String()] = now.Add(p.cooldown/2 + jitter)
	p.mu.Unlock()
}

type Transport struct {
	base     *http.Transport
	pool     *Pool
	byProxy  map[string]*http.Transport
	mu       sync.Mutex
	modeOnly bool
}

func NewTransport(base *http.Transport, pool *Pool, onlyProxy bool) *Transport {
	if base == nil {
		base = http.DefaultTransport.(*http.Transport).Clone()
	}
	return &Transport{
		base:     base,
		pool:     pool,
		byProxy:  make(map[string]*http.Transport),
		modeOnly: onlyProxy,
	}
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.pool == nil {
		if t.modeOnly {
			return nil, errors.New("proxy required but no pool configured")
		}
		return t.base.RoundTrip(req)
	}

	now := time.Now()
	t.pool.mu.Lock()
	proxyURL := t.pool.pickLocked(now)
	nextAfter := t.pool.nextUsableAfterLocked(now)
	t.pool.mu.Unlock()

	if !nextAfter.IsZero() && nextAfter.After(now) {
		wait := time.Until(nextAfter)
		t := time.NewTimer(wait)
		select {
		case <-req.Context().Done():
			t.Stop()
			return nil, req.Context().Err()
		case <-t.C:
		}
	}

	rt := t.transportFor(proxyURL)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.pool.MarkMaybeBad(proxyURL)
		return resp, err
	}
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusServiceUnavailable {
		t.pool.MarkBad(proxyURL)
	}
	return resp, nil
}

func (t *Transport) transportFor(proxyURL *url.URL) *http.Transport {
	key := proxyURL.String()
	t.mu.Lock()
	defer t.mu.Unlock()
	if tr, ok := t.byProxy[key]; ok {
		return tr
	}
	tr := t.base.Clone()
	tr.Proxy = http.ProxyURL(proxyURL)
	t.byProxy[key] = tr
	return tr
}

