package usecase

import (
	"context"
	"testing"

	"github.com/butbeautifulv/veil/pkg/harvest"
)

type mockPub struct {
	kind string
	key  string
}

func (m *mockPub) Publish(ctx context.Context, kind, contentKey string, payload any) error {
	m.kind = kind
	m.key = contentKey
	return nil
}

func TestPublishOSV_kindAndKey(t *testing.T) {
	pub := &mockPub{}
	pl := harvest.SBOMOSVRaw{CVE: "CVE-2024-0001", OSVID: "OSV-1", RawJSON: `{}`}
	if err := pub.Publish(context.Background(), harvest.KindSBOMOSVJSON, "sbom:osv:CVE-2024-0001", pl); err != nil {
		t.Fatal(err)
	}
	if pub.kind != harvest.KindSBOMOSVJSON {
		t.Fatalf("kind = %q", pub.kind)
	}
	if pub.key != "sbom:osv:CVE-2024-0001" {
		t.Fatalf("key = %q", pub.key)
	}
}
