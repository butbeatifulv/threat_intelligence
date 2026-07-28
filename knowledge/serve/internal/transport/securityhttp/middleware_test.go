package securityhttp

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/butbeautifulv/veil/knowledge/serve/internal/config"
)

func TestHarden_securityHeaders(t *testing.T) {
	sec := config.SecurityConfig{}
	h := Harden(sec, 0, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("missing nosniff")
	}
}

func TestHarden_corsDenied(t *testing.T) {
	sec := config.SecurityConfig{CORSAllowedOrigins: []string{"https://allowed.example"}}
	h := Harden(sec, 0, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.example")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestHarden_bodyLimit(t *testing.T) {
	sec := config.SecurityConfig{}
	var readErr error
	h := Harden(sec, 10, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, readErr = io.ReadAll(r.Body)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(make([]byte, 32)))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if readErr == nil {
		t.Fatal("expected body limit error")
	}
}
