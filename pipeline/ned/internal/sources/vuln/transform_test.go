package vuln

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/butbeautifulv/veil/pkg/harvest"
	"github.com/butbeautifulv/veil/pkg/vuln/domain"
)

func TestTransform_NVDPage_CWEAndCPE(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "pkg", "nvd", "parse", "testdata", "nvd_page_min.json"))
	if err != nil {
		t.Fatal(err)
	}
	page := harvest.VulnNVDPage{RawJSON: string(raw)}
	payload, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	env := &harvest.Envelope{Kind: harvest.KindVulnNVDPage, Payload: payload}
	out, err := Transform(context.Background(), env)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("envelopes: got %d want 2", len(out))
	}
	var v domain.Vulnerability
	if err := json.Unmarshal(out[0].Payload, &v); err != nil {
		t.Fatal(err)
	}
	if len(v.CWE) == 0 {
		t.Fatalf("expected CWE on first CVE, got %#v", v)
	}
	if len(v.CPEs) == 0 {
		t.Fatalf("expected CPEs on first CVE, got %#v", v)
	}
}
