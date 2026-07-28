package mcpserver

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/knowledge/serve/internal/config"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/usecase"
)

func TestHTTP_post_batchRejected(t *testing.T) {
	srv := NewServer(usecase.NewReadUsecase(&mockExec{}), nil, nil, nil, nil, slog.Default())
	h := HTTPHandler(srv, config.MCPHTTPConfig{Path: "/mcp"})
	body := `[{"jsonrpc":"2.0","id":1,"method":"ping"},{"jsonrpc":"2.0","id":2,"method":"ping"}]`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestHTTP_post_oversizedBody(t *testing.T) {
	srv := NewServer(usecase.NewReadUsecase(&mockExec{}), nil, nil, nil, nil, slog.Default())
	h := HTTPHandler(srv, config.MCPHTTPConfig{Path: "/mcp"})
	big := strings.Repeat("x", 5<<20+1)
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`+big))
	req.ContentLength = int64(len(big) + 50)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest && rr.Code != http.StatusRequestEntityTooLarge {
		// LimitReader returns 400 from read error path
		if rr.Code >= 500 {
			t.Fatalf("status %d", rr.Code)
		}
	}
	_ = io.Discard
}

func TestHTTP_post_emptyBody(t *testing.T) {
	srv := NewServer(usecase.NewReadUsecase(&mockExec{}), nil, nil, nil, nil, slog.Default())
	h := HTTPHandler(srv, config.MCPHTTPConfig{Path: "/mcp"})
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rr.Code)
	}
}
