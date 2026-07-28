package ingest

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
)

func TestApplyEnvelope_unknownKind(t *testing.T) {
	err := ApplyEnvelope(context.Background(), nil, &commit.Envelope{Kind: "unknown"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestApplyEnvelope_toolRunPayload(t *testing.T) {
	payload, _ := json.Marshal(commit.EngageToolRunPayload{
		Tool: "nuclei", Target: "https://example.com", Success: true, At: "2026-05-16T00:00:00Z",
	})
	var p commit.EngageToolRunPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		t.Fatal(err)
	}
	if p.Tool != "nuclei" || p.Target == "" {
		t.Fatalf("payload %+v", p)
	}
}
