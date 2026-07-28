package serve

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	discbrowser "github.com/butbeautifulv/veil/discovery/pkg/browser"
	"github.com/butbeautifulv/veil/pkg/harvest"
)

func TestHandler_inspect_publishesHarvest(t *testing.T) {
	sidecar := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(discbrowser.InspectResult{
			Success:   true,
			Timestamp: "2026-01-01T00:00:00Z",
			PageInfo:  map[string]any{"url": "https://example.com"},
		})
	}))
	defer sidecar.Close()

	var published bool
	pub := harvestPubStub{fn: func(_ context.Context, kind, _ string, _ any) error {
		if kind == harvest.KindBrowserInspectRaw {
			published = true
		}
		return nil
	}}

	cfg := Config{
		Sidecar:    &discbrowser.Sidecar{BaseURL: sidecar.URL, Client: sidecar.Client()},
		HarvestPub: pub,
	}
	mux := Handler(cfg)

	body := []byte(`{"url":"https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/inspect", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if !published {
		t.Fatal("expected harvest publish")
	}
}

type harvestPubStub struct {
	fn func(context.Context, string, string, any) error
}

func (h harvestPubStub) Publish(ctx context.Context, kind, contentKey string, payload any) error {
	return h.fn(ctx, kind, contentKey, payload)
}
