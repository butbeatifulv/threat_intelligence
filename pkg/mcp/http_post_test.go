package mcp

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTP_post_emptyBody(t *testing.T) {
	h := HTTPHandler(stubProc{}, HTTPConfig{Path: "/mcp"})
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestHTTP_post_batchRejected(t *testing.T) {
	h := HTTPHandler(stubProc{}, HTTPConfig{Path: "/mcp"})
	body := `[{"jsonrpc":"2.0","id":1,"method":"ping"},{"jsonrpc":"2.0","id":2,"method":"pong"}]`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestHTTP_post_notificationOnly(t *testing.T) {
	h := HTTPHandler(stubProc{}, HTTPConfig{Path: "/mcp"})
	body := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("status %d", rr.Code)
	}
}

func TestHTTP_getMethodNotAllowed(t *testing.T) {
	h := HTTPHandler(stubProc{}, HTTPConfig{Path: "mcp", Service: "svc"})
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d", rr.Code)
	}
}
