package mcpserver

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authmw "github.com/butbeautifulv/veil/knowledge/serve/internal/auth/middleware"
	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/auth/keycloak"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/config"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/usecase"
)

func TestHTTP_initialize_json(t *testing.T) {
	uc := usecase.NewReadUsecase(&mockExec{})
	srv := NewServer(uc, nil, nil, nil, nil, slog.Default())
	cfg := config.MCPHTTPConfig{Path: "/mcp"}
	h := HTTPHandler(srv, cfg)

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type %q", ct)
	}
	var msg rpcMessage
	if err := json.Unmarshal(rr.Body.Bytes(), &msg); err != nil {
		t.Fatal(err)
	}
	pv, _ := msg.Result.(map[string]any)["protocolVersion"].(string)
	if pv != protocolVersionHTTP {
		t.Fatalf("protocol %q", pv)
	}
}

func TestHTTP_health(t *testing.T) {
	srv := NewServer(usecase.NewReadUsecase(&mockExec{}), nil, nil, nil, nil, slog.Default())
	h := HTTPHandler(srv, config.MCPHTTPConfig{Path: "/mcp"})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestHTTP_get_mcp_405(t *testing.T) {
	srv := NewServer(usecase.NewReadUsecase(&mockExec{}), nil, nil, nil, nil, slog.Default())
	h := HTTPHandler(srv, config.MCPHTTPConfig{Path: "/mcp"})
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestHTTP_auth_required(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://keycloak.example/realms/veil"
	aud := "veil-api"
	v := keycloak.NewStaticVerifier(issuer, aud, "veil-api", &key.PublicKey)
	stack := auth.NewStack(v, auth.Config{
		Enabled:     true,
		RBACEnabled: true,
		RoleReader:  "veil-reader",
	})

	uc := usecase.NewReadUsecase(&mockExec{})
	srv := NewServer(uc, nil, nil, nil, stack, slog.Default())
	cfg := config.MCPHTTPConfig{Path: "/mcp"}
	h := authmw.Auth(stack, true, config.SecurityConfig{}, HTTPHandler(srv, cfg))

	body := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"ti_health","arguments":{}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rr.Code)
	}

	tok, _ := keycloak.SignTestToken(key, issuer, aud, "u1", []string{"veil-reader"}, time.Hour)
	req2 := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	req2.Header.Set("Authorization", "Bearer "+tok)
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr2.Code, rr2.Body.String())
	}
}

func TestHTTP_tools_call_via_bearer(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://keycloak.example/realms/veil"
	aud := "veil-api"
	v := keycloak.NewStaticVerifier(issuer, aud, "veil-api", &key.PublicKey)
	stack := auth.NewStack(v, auth.Config{Enabled: true, RBACEnabled: false})
	tok, _ := keycloak.SignTestToken(key, issuer, aud, "u1", nil, time.Hour)

	uc := usecase.NewReadUsecase(&mockExec{})
	srv := NewServer(uc, nil, nil, nil, stack, slog.Default())
	h := authmw.Auth(stack, false, config.SecurityConfig{}, HTTPHandler(srv, config.MCPHTTPConfig{Path: "/mcp"}))

	body := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"ti_health","arguments":{}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	_ = context.Background()
}
