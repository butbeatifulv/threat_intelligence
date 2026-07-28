package broker

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Addr        string
	Token       string
	CheckEvery  time.Duration
	HTTPTimeout time.Duration
	TestURL     string
}

type Broker struct {
	cfg    Config
	logger *slog.Logger

	mu     sync.Mutex
	items  map[string]*proxyItem
	order  []string
	leases map[string]time.Time // key -> until
	next   int
}

type proxyItem struct {
	URL        *url.URL
	AddedAt    time.Time
	LastOK     time.Time
	LastFail   time.Time
	FailCount  int
	DisabledTo time.Time
}

func New(cfg Config, logger *slog.Logger) *Broker {
	if cfg.CheckEvery <= 0 {
		cfg.CheckEvery = 45 * time.Second
	}
	if cfg.HTTPTimeout <= 0 {
		cfg.HTTPTimeout = 10 * time.Second
	}
	if strings.TrimSpace(cfg.TestURL) == "" {
		cfg.TestURL = "https://example.com/"
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Broker{
		cfg:    cfg,
		logger: logger,
		items:  make(map[string]*proxyItem),
		leases: make(map[string]time.Time),
	}
}

func (b *Broker) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", b.handleHealthz)
	mux.HandleFunc("POST /v1/proxies", b.auth(b.handleAddProxies))
	mux.HandleFunc("GET /v1/proxies", b.auth(b.handleListProxies))
	mux.HandleFunc("POST /v1/report", b.auth(b.handleReport))
	mux.HandleFunc("GET /v1/lease", b.auth(b.handleLease))
	return mux
}

func (b *Broker) auth(next http.HandlerFunc) http.HandlerFunc {
	// If Token is empty, run open (useful for local dev).
	if strings.TrimSpace(b.cfg.Token) == "" {
		return next
	}
	return func(w http.ResponseWriter, r *http.Request) {
		got := strings.TrimSpace(r.Header.Get("Authorization"))
		const prefix = "Bearer "
		if strings.HasPrefix(got, prefix) {
			got = strings.TrimSpace(strings.TrimPrefix(got, prefix))
		}
		if subtle.ConstantTimeCompare([]byte(got), []byte(b.cfg.Token)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (b *Broker) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

type addProxiesRequest struct {
	Proxies []string `json:"proxies"`
}

type addProxiesResponse struct {
	Added   int      `json:"added"`
	Invalid []string `json:"invalid"`
}

func (b *Broker) handleAddProxies(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var req addProxiesRequest
	if err := json.Unmarshal(body, &req); err != nil {
		// allow plain-text body: one proxy per line
		lines := strings.Split(string(body), "\n")
		for _, ln := range lines {
			ln = strings.TrimSpace(ln)
			if ln != "" {
				req.Proxies = append(req.Proxies, ln)
			}
		}
	}

	added := 0
	var invalid []string
	now := time.Now()

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, s := range req.Proxies {
		u, perr := parseProxyURL(s)
		if perr != nil {
			invalid = append(invalid, s)
			continue
		}
		key := u.String()
		if _, ok := b.items[key]; ok {
			continue
		}
		b.items[key] = &proxyItem{URL: u, AddedAt: now}
		b.order = append(b.order, key)
		added++
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(addProxiesResponse{Added: added, Invalid: invalid})
}

type proxyView struct {
	URL        string    `json:"url"`
	AddedAt    time.Time `json:"added_at"`
	LastOK     time.Time `json:"last_ok,omitempty"`
	LastFail   time.Time `json:"last_fail,omitempty"`
	FailCount  int       `json:"fail_count"`
	DisabledTo time.Time `json:"disabled_to,omitempty"`
	LeasedTo   time.Time `json:"leased_to,omitempty"`
}

func (b *Broker) handleListProxies(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	b.mu.Lock()
	out := make([]proxyView, 0, len(b.order))
	for _, key := range b.order {
		it := b.items[key]
		if it == nil {
			continue
		}
		out = append(out, proxyView{
			URL:        key,
			AddedAt:    it.AddedAt,
			LastOK:     it.LastOK,
			LastFail:   it.LastFail,
			FailCount:  it.FailCount,
			DisabledTo: it.DisabledTo,
			LeasedTo:   maxTime(b.leases[key], time.Time{}),
		})
	}
	// cleanup expired leases opportunistically
	for key, until := range b.leases {
		if !until.IsZero() && until.Before(now) {
			delete(b.leases, key)
		}
	}
	b.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"proxies": out})
}

type leaseResponse struct {
	Proxies []string `json:"proxies"`
}

func (b *Broker) handleLease(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	n := parseInt(q.Get("n"), 1, 50, 5)
	ttl := parseDur(q.Get("ttl"), 5*time.Second, 30*time.Minute, 2*time.Minute)
	now := time.Now()

	b.mu.Lock()
	defer b.mu.Unlock()

	// cleanup leases
	for key, until := range b.leases {
		if !until.IsZero() && until.Before(now) {
			delete(b.leases, key)
		}
	}

	var out []string
	total := len(b.order)
	for tries := 0; tries < total && len(out) < n; tries++ {
		if total == 0 {
			break
		}
		key := b.order[b.next%total]
		b.next = (b.next + 1) % total

		it := b.items[key]
		if it == nil {
			continue
		}
		if it.DisabledTo.After(now) {
			continue
		}
		if until, leased := b.leases[key]; leased && until.After(now) {
			continue
		}
		// Prefer proxies that have been OK at least once.
		if it.LastOK.IsZero() && it.FailCount >= 2 {
			continue
		}
		b.leases[key] = now.Add(ttl)
		out = append(out, key)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(leaseResponse{Proxies: out})
}

type reportRequest struct {
	Proxy  string `json:"proxy"`
	Reason string `json:"reason"`
	// Optional: seconds to disable the proxy for.
	DisableSeconds int `json:"disable_seconds"`
}

func (b *Broker) handleReport(w http.ResponseWriter, r *http.Request) {
	var req reportRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	req.Proxy = strings.TrimSpace(req.Proxy)
	if req.Proxy == "" {
		http.Error(w, "proxy required", http.StatusBadRequest)
		return
	}
	_, err := parseProxyURL(req.Proxy)
	if err != nil {
		http.Error(w, "invalid proxy", http.StatusBadRequest)
		return
	}
	disableFor := time.Duration(req.DisableSeconds) * time.Second
	if disableFor <= 0 {
		disableFor = 2 * time.Minute
	}

	now := time.Now()
	b.mu.Lock()
	it := b.items[req.Proxy]
	if it != nil {
		it.LastFail = now
		it.FailCount++
		it.DisabledTo = now.Add(disableFor)
	}
	delete(b.leases, req.Proxy)
	b.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (b *Broker) RunHealthLoop(ctx context.Context) {
	t := time.NewTicker(b.cfg.CheckEvery)
	defer t.Stop()

	// do a quick pass soon after start
	b.healthPass(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			b.healthPass(ctx)
		}
	}
}

func (b *Broker) healthPass(ctx context.Context) {
	// Snapshot keys to avoid holding the lock across network I/O.
	b.mu.Lock()
	keys := append([]string(nil), b.order...)
	b.mu.Unlock()

	if len(keys) == 0 {
		return
	}

	testURL := b.cfg.TestURL
	timeout := b.cfg.HTTPTimeout
	now := time.Now()

	for _, key := range keys {
		select {
		case <-ctx.Done():
			return
		default:
		}
		u, _ := url.Parse(key)
		ok := probeOnce(ctx, u, testURL, timeout)

		b.mu.Lock()
		it := b.items[key]
		if it == nil {
			b.mu.Unlock()
			continue
		}
		if ok {
			it.LastOK = now
			it.FailCount = 0
			if it.DisabledTo.Before(now) {
				it.DisabledTo = time.Time{}
			}
		} else {
			it.LastFail = now
			it.FailCount++
			// progressive backoff disable
			d := time.Duration(minInt(6, it.FailCount)) * 30 * time.Second
			it.DisabledTo = now.Add(d)
			delete(b.leases, key)
		}
		b.mu.Unlock()
	}
}

func probeOnce(ctx context.Context, proxyURL *url.URL, testURL string, timeout time.Duration) bool {
	base := http.DefaultTransport.(*http.Transport).Clone()
	base.Proxy = http.ProxyURL(proxyURL)
	base.TLSHandshakeTimeout = 8 * time.Second
	base.DialContext = (&net.Dialer{Timeout: 6 * time.Second, KeepAlive: 30 * time.Second}).DialContext
	c := &http.Client{Timeout: timeout, Transport: base}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, testURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "proxybroker/1.0")
	resp, err := c.Do(req)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	// Any 2xx/3xx indicates we can reach the outside world.
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func parseProxyURL(s string) (*url.URL, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("empty")
	}
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("scheme must be http/https")
	}
	if u.Host == "" {
		return nil, errors.New("missing host")
	}
	return u, nil
}

func parseInt(s string, minV, maxV, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return def
		}
		n = n*10 + int(ch-'0')
	}
	if n < minV {
		return minV
	}
	if n > maxV {
		return maxV
	}
	return n
}

func parseDur(s string, minV, maxV, def time.Duration) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return def
	}
	if d < minV {
		return minV
	}
	if d > maxV {
		return maxV
	}
	return d
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

