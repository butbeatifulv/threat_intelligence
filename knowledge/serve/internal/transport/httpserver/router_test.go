package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/butbeautifulv/veil/knowledge/serve/internal/usecase"
)

type mockReadExec struct{}

func (mockReadExec) ExecRead(ctx context.Context, fn func(tx driver.ManagedTransaction) (any, error)) (any, error) {
	return nil, nil
}

func TestRouter_categories_and_health(t *testing.T) {
	uc := usecase.NewReadUsecase(mockReadExec{})
	mux := http.NewServeMux()
	Register(mux, uc, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("health: %d", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/v1/categories", nil)
	rr2 := httptest.NewRecorder()
	mux.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("categories: %d %s", rr2.Code, rr2.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(rr2.Body.Bytes(), &body)
	cats, ok := body["categories"].([]any)
	if !ok {
		t.Fatalf("body: %v", body)
	}
	foundEngage := false
	for _, c := range cats {
		m, _ := c.(map[string]any)
		if m["id"] == "engage" {
			foundEngage = true
			break
		}
	}
	if !foundEngage {
		t.Fatalf("engage category missing: %v", cats)
	}
}
