package mcpserver

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/knowledge/serve/internal/config"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/usecase"
)

func TestHTTP_initialize_sse(t *testing.T) {
	srv := NewServer(usecase.NewReadUsecase(&mockExec{}), nil, nil, nil, nil, slog.Default())
	cfg := config.MCPHTTPConfig{Path: "/mcp", PreferSSE: true}
	h := HTTPHandler(srv, cfg)

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("content-type %q", ct)
	}
	out := rr.Body.String()
	if !strings.Contains(out, "event: message") || !strings.Contains(out, "data: ") {
		t.Fatalf("sse body: %s", out)
	}
	if !strings.Contains(out, `"protocolVersion"`) {
		t.Fatalf("missing protocol in sse: %s", out)
	}
}

func TestWantsSSE_acceptHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Accept", "application/json, text/event-stream")
	if !wantsSSE(req, false) {
		t.Fatal("expected sse from accept")
	}
	req2 := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req2.Header.Set("Accept", "application/json")
	if wantsSSE(req2, false) {
		t.Fatal("expected json only")
	}
	if !wantsSSE(req2, true) {
		t.Fatal("expected prefer sse")
	}
}
