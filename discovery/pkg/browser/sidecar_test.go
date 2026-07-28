package browser

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/butbeautifulv/veil/pkg/harvest"
)

func TestSidecar_Inspect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/inspect" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"page_info": map[string]any{
				"title": "Example",
				"forms": []any{map[string]any{"action": "/login"}},
			},
			"security_analysis": map[string]any{"security_score": 80},
		})
	}))
	defer srv.Close()

	svc := &Sidecar{BaseURL: srv.URL, Client: srv.Client()}
	out := svc.Inspect(context.Background(), InspectRequest{URL: "https://example.com"})
	if !out.Success {
		t.Fatalf("success false: %s", out.Error)
	}
	if len(out.Forms) != 1 {
		t.Fatalf("forms: %#v", out.Forms)
	}
}

func TestCatalogProxy_Inspect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"security_analysis": map[string]any{"security_score": 90},
		})
	}))
	defer srv.Close()

	proxy := NewCatalogProxy(srv.URL)
	res := proxy.Inspect(context.Background(), InspectRequest{URL: "https://example.com"})
	if res.ExitCode != 0 {
		t.Fatalf("res: %+v", res)
	}
}

func TestIsBrowserBinary(t *testing.T) {
	if !IsBrowserBinary("browser") {
		t.Fatal("expected browser")
	}
}

func TestPublishInspect(t *testing.T) {
	var gotKind string
	pub := harvestPubStub{fn: func(_ context.Context, kind, _ string, payload any) error {
		gotKind = kind
		if _, ok := payload.(harvest.BrowserInspectRaw); !ok {
			t.Fatalf("payload type %T", payload)
		}
		return nil
	}}
	err := PublishInspect(context.Background(), pub, "https://example.com", InspectResult{Success: true, Timestamp: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if gotKind != harvest.KindBrowserInspectRaw {
		t.Fatalf("kind %q", gotKind)
	}
}

type harvestPubStub struct {
	fn func(context.Context, string, string, any) error
}

func (h harvestPubStub) Publish(ctx context.Context, kind, contentKey string, payload any) error {
	return h.fn(ctx, kind, contentKey, payload)
}
