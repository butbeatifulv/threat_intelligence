package scrapepub

import (
	"context"
	"testing"

	"github.com/butbeautifulv/veil/pkg/harvest"
)

type fakeRaw struct {
	source  string
	kind    string
	key     string
	payload any
}

func (f *fakeRaw) Publish(_ context.Context, kind, contentKey string, payload any) error {
	f.kind = kind
	f.key = contentKey
	f.payload = payload
	return nil
}

func TestNewBase_publish(t *testing.T) {
	raw := &fakeRaw{source: harvest.SourceTI}
	b := NewBase(raw)
	if err := b.Raw.Publish(context.Background(), harvest.KindTIIoCRaw, "ti:ioc:ip:1.2.3.4", nil); err != nil {
		t.Fatal(err)
	}
	if raw.kind != harvest.KindTIIoCRaw || raw.key != "ti:ioc:ip:1.2.3.4" {
		t.Fatalf("got kind=%q key=%q", raw.kind, raw.key)
	}
}
