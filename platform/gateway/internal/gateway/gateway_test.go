package gateway

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/butbeautifulv/veil/platform/gateway/internal/config"
)

func TestNewMux_routesAndCompositeHealth(t *testing.T) {
	graph := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "service": "veil-api"})
		case "/v1/categories":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("graph-ok"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer graph.Close()

	engage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "service": "engage-api"})
		case "/api/tools/list":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("engage-ok"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer engage.Close()

	cfg := config.Config{
		GraphAPIURL:  mustURL(t, graph.URL),
		EngageAPIURL: mustURL(t, engage.URL),
	}
	mux := NewMux(cfg)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	t.Run("proxy graph v1", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/v1/categories")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK || string(body) != "graph-ok" {
			t.Fatalf("status=%d body=%q", resp.StatusCode, body)
		}
	})

	t.Run("proxy engage api", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/api/tools/list")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK || string(body) != "engage-ok" {
			t.Fatalf("status=%d body=%q", resp.StatusCode, body)
		}
	})

	t.Run("composite health ok", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/health")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var body map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK || body["ok"] != true {
			t.Fatalf("status=%d body=%v", resp.StatusCode, body)
		}
	})
}

func TestCompositeHealth_upstreamDown(t *testing.T) {
	graph := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer graph.Close()

	engage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer engage.Close()

	cfg := config.Config{
		GraphAPIURL:  mustURL(t, graph.URL),
		EngageAPIURL: mustURL(t, engage.URL),
	}
	srv := httptest.NewServer(NewMux(cfg))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status=%d want 503", resp.StatusCode)
	}
}

func mustURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return u
}
