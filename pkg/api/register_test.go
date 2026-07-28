package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterHealth(t *testing.T) {
	mux := http.NewServeMux()
	RegisterHealth(mux, "test-svc", map[string]any{"n": 3})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["service"] != "test-svc" || body["n"].(float64) != 3 {
		t.Fatalf("body %v", body)
	}
}

func TestPostJSON(t *testing.T) {
	mux := http.NewServeMux()
	PostJSON(mux, "POST /echo", func(r *http.Request, body map[string]any) (any, int) {
		return map[string]any{"echo": body["x"]}, http.StatusCreated
	})

	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(`{"x":"hi"}`))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
}

func TestPostJSON_emptyBody(t *testing.T) {
	mux := http.NewServeMux()
	PostJSON(mux, "POST /noop", func(r *http.Request, body map[string]any) (any, int) {
		if body == nil || len(body) != 0 {
			t.Fatalf("body %v", body)
		}
		return map[string]any{"ok": true}, http.StatusOK
	})
	req := httptest.NewRequest(http.MethodPost, "/noop", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
}
