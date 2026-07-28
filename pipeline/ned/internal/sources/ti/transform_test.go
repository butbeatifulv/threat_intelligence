package ti

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
	tidomain "github.com/butbeautifulv/veil/pkg/ti/domain"
	tinormalize "github.com/butbeautifulv/veil/pkg/ti/normalize"
)

func TestTransform_IoCRaw_normalizesInPipeline(t *testing.T) {
	raw := tidomain.IOC{
		Type:   tidomain.IOCURL,
		Value:  "HTTPS://Evil.COM/path",
		Source: "urlhaus",
	}
	payload, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	env := &harvest.Envelope{
		SchemaVersion: harvest.CurrentSchemaVersion,
		Source:        harvest.SourceTI,
		Kind:          harvest.KindTIIoCRaw,
		ContentKey:    "ti:ioc:url:https://evil.com/path",
		ScrapedAt:     "2026-05-15T12:00:00Z",
		Payload:       payload,
	}
	out, err := Transform(context.Background(), env)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("len(out) = %d", len(out))
	}
	if out[0].Kind != commit.KindTIIoC {
		t.Fatalf("kind = %s", out[0].Kind)
	}
	var ni tidomain.IOC
	if err := json.Unmarshal(out[0].Payload, &ni); err != nil {
		t.Fatal(err)
	}
	wantID := tinormalize.CanonicalID(ni)
	if out[0].IdempotencyKey != commit.TIIoCIdempotencyKey(wantID) {
		t.Fatalf("idempotency %q want %q", out[0].IdempotencyKey, commit.TIIoCIdempotencyKey(wantID))
	}
	if ni.Value != "https://evil.com/path" {
		t.Fatalf("normalized value = %q", ni.Value)
	}
}
