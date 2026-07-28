package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/auth/keycloak"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/config"
)

func TestAuth_middleware(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://keycloak.example/realms/veil"
	aud := "veil-api"
	v := keycloak.NewStaticVerifier(issuer, aud, "veil-api", &key.PublicKey)
	cfg := auth.Config{
		Enabled:     true,
		RBACEnabled: true,
		RoleReader:  "veil-reader",
	}
	stack := auth.NewStack(v, cfg)
	sec := config.SecurityConfig{}

	okTok, _ := keycloak.SignTestToken(key, issuer, aud, "u1", []string{"veil-reader"}, time.Hour)
	badTok, _ := keycloak.SignTestToken(key, issuer, aud, "u2", []string{"nope"}, time.Hour)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /v1/categories", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := Auth(stack, false, sec, mux)

	t.Run("health open", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status %d", rr.Code)
		}
	})

	t.Run("no token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/categories", nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status %d", rr.Code)
		}
	})

	t.Run("valid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/categories", nil)
		req.Header.Set("Authorization", "Bearer "+okTok)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("forbidden role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/categories", nil)
		req.Header.Set("Authorization", "Bearer "+badTok)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("status %d", rr.Code)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		open := auth.NewStack(nil, auth.Config{Enabled: false})
		h2 := Auth(open, false, sec, mux)
		req := httptest.NewRequest(http.MethodGet, "/v1/categories", nil)
		rr := httptest.NewRecorder()
		h2.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("status %d", rr.Code)
		}
	})
}

func TestAuth_strictMCP(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://keycloak.example/realms/veil"
	v := keycloak.NewStaticVerifier(issuer, "veil-api", "veil-api", &key.PublicKey)
	stack := auth.NewStack(v, auth.Config{Enabled: true, RBACEnabled: false})
	tok, _ := keycloak.SignTestToken(key, issuer, "veil-api", "u1", nil, time.Hour)
	sec := config.SecurityConfig{}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("POST /mcp", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := Auth(stack, true, sec, mux)

	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusUnauthorized {
		t.Fatalf("strict: status %d", rr2.Code)
	}
}

func TestAuth_subjectInContext(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://keycloak.example/realms/veil"
	v := keycloak.NewStaticVerifier(issuer, "veil-api", "veil-api", &key.PublicKey)
	stack := auth.NewStack(v, auth.Config{Enabled: true, RBACEnabled: false})
	tok, _ := keycloak.SignTestToken(key, issuer, "veil-api", "u1", nil, time.Hour)

	var gotSub string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/kinds", func(w http.ResponseWriter, r *http.Request) {
		sub, ok := auth.SubjectFromContext(r.Context())
		if !ok {
			t.Error("no subject")
			return
		}
		gotSub = sub.Sub
	})
	req := httptest.NewRequest(http.MethodGet, "/v1/kinds", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	Auth(stack, false, config.SecurityConfig{}, mux).ServeHTTP(httptest.NewRecorder(), req)
	if gotSub != "u1" {
		t.Fatalf("sub %q", gotSub)
	}
}
