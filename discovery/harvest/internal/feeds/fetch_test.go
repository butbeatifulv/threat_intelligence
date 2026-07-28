package feeds

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/ledger"
)

type memLedger struct {
	sha       map[string]string
	fetched   map[string]time.Time
	policy    map[string]ledger.FetchPolicy
	recorded  []string
	forceNext bool
}

func newMemLedger() *memLedger {
	return &memLedger{
		sha:     make(map[string]string),
		fetched: make(map[string]time.Time),
		policy:  make(map[string]ledger.FetchPolicy),
	}
}

func (m *memLedger) ShouldFetch(_ context.Context, key string, policy ledger.FetchPolicy, minRefetch time.Duration, force bool) (bool, error) {
	if force || m.forceNext {
		return true, nil
	}
	if policy == ledger.PolicyStatic {
		_, ok := m.fetched[key]
		return !ok, nil
	}
	t, ok := m.fetched[key]
	if !ok {
		return true, nil
	}
	if policy == ledger.PolicyDaily {
		return time.Since(t) >= 24*time.Hour, nil
	}
	if minRefetch <= 0 {
		return true, nil
	}
	return time.Since(t) >= minRefetch, nil
}

func (m *memLedger) GetContentSHA(_ context.Context, key string) (string, error) {
	return m.sha[key], nil
}

func (m *memLedger) RecordFetch(_ context.Context, key, _, _ string, policy ledger.FetchPolicy, contentSHA256 string) error {
	m.fetched[key] = time.Now().UTC()
	m.sha[key] = contentSHA256
	m.policy[key] = policy
	m.recorded = append(m.recorded, key)
	return nil
}

func TestFetchIfDue_unchangedOnSecondFetch(t *testing.T) {
	body := []byte(`{"ok":true}`)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	dir := t.TempDir()
	c := NewClient(dir, nil)
	led := newMemLedger()
	key := "test:resource"

	res1, err := FetchIfDue(context.Background(), c, led, key, "test", srv.URL, ledger.PolicyPeriodic, "t/res.json", func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, srv.URL, nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	if res1.Unchanged {
		t.Fatal("first fetch should not be unchanged")
	}

	res2, err := FetchIfDue(context.Background(), c, led, key, "test", srv.URL, ledger.PolicyPeriodic, "t/res.json", func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, srv.URL, nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res2.Unchanged {
		t.Fatal("second fetch with same body should be unchanged")
	}
}

func TestFetchIfDue_skippedUsesCacheUnchanged(t *testing.T) {
	body := []byte("cached-body")
	dir := t.TempDir()
	c := NewClient(dir, nil)
	_ = c.WriteCache("t/c.csv", body)

	sum := sha256.Sum256(body)
	sha := hex.EncodeToString(sum[:])
	led := newMemLedger()
	led.sha["ti:urlhaus"] = sha
	led.fetched["ti:urlhaus"] = time.Now().UTC()

	res, err := FetchIfDue(context.Background(), c, led, "ti:urlhaus", "ti", "http://example.invalid/", ledger.PolicyDaily, "t/c.csv", func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, "http://example.invalid/", nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Skipped || !res.Unchanged {
		t.Fatalf("Skipped=%v Unchanged=%v", res.Skipped, res.Unchanged)
	}
}

func TestFetchIfDue_postUnchangedOnSecondFetch(t *testing.T) {
	body := []byte(`{"query":"get_recent"}`)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	dir := t.TempDir()
	c := NewClient(dir, nil)
	led := newMemLedger()
	key := "ti:malwarebazaar:recent"
	build := func() (*http.Request, error) {
		req, err := http.NewRequest(http.MethodPost, srv.URL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	}

	res1, err := FetchIfDue(context.Background(), c, led, key, "ti", srv.URL, ledger.PolicyDaily, "ti/mb.json", build)
	if err != nil {
		t.Fatal(err)
	}
	if res1.Unchanged {
		t.Fatal("first POST fetch should not be unchanged")
	}

	res2, err := FetchIfDue(context.Background(), c, led, key, "ti", srv.URL, ledger.PolicyDaily, "ti/mb.json", build)
	if err != nil {
		t.Fatal(err)
	}
	if !res2.Unchanged {
		t.Fatal("second POST fetch with same body should be unchanged")
	}
}

func TestFetchIfDue_ledgerSkipCacheMissRefetches(t *testing.T) {
	body := []byte(`{"refetch":true}`)
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	dir := t.TempDir()
	c := NewClient(dir, nil)
	led := newMemLedger()
	key := "lola:mitre:enterprise-stix"
	sum := sha256.Sum256(body)
	sha := hex.EncodeToString(sum[:])
	led.sha[key] = sha
	led.fetched[key] = time.Now().UTC()

	res, err := FetchIfDue(context.Background(), c, led, key, "lola", srv.URL, ledger.PolicyStatic, "mitre/enterprise-attack.json", func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, srv.URL, nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Skipped || len(res.Body) == 0 {
		t.Fatalf("expected refetch body, Skipped=%v len=%d", res.Skipped, len(res.Body))
	}
	if hits != 1 {
		t.Fatalf("HTTP hits = %d, want 1", hits)
	}
	if _, hit := c.ReadCache("mitre/enterprise-attack.json"); !hit {
		t.Fatal("expected cache write after refetch")
	}
}
