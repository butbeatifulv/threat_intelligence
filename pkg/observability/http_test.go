package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWrapHandler_metricsAndHealth(t *testing.T) {
	inner := httptest.NewServer(httpHandlerStub())
	defer inner.Close()

	// Use a minimal stub via WrapHandler on a handler that serves /health
	h := WrapHandler("test-svc", httpHandlerStub())

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("health status = %d", rec.Code)
	}

	req2 := httptest.NewRequest("GET", "/metrics", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != 200 {
		t.Fatalf("metrics status = %d", rec2.Code)
	}
	if !contains(rec2.Body.String(), "veil_http_requests_total") {
		t.Fatalf("expected veil_http_requests_total in metrics body")
	}
}

func httpHandlerStub() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	return mux
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
