package commit

import (
	"encoding/json"
	"testing"
)

func TestEngageFindingEnvelope(t *testing.T) {
	payload, _ := json.Marshal(EngageFindingPayload{
		Tool: "nuclei", Target: "https://example.com", Title: "xss", Severity: "high", Description: "reflected",
	})
	env := &Envelope{
		SchemaVersion:  CurrentSchemaVersion,
		Source:         SourceEngage,
		Kind:           KindEngageFinding,
		IdempotencyKey: EngageFindingIdempotencyKey("nuclei", "https://example.com", "xss"),
		Payload:        payload,
	}
	if err := env.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestEngageToolRunEnvelope(t *testing.T) {
	payload, _ := json.Marshal(EngageToolRunPayload{
		Tool: "nuclei", Target: "https://example.com", Subject: "svc", Success: true, At: "2026-05-16T00:00:00Z",
	})
	env := &Envelope{
		SchemaVersion:  CurrentSchemaVersion,
		Source:         SourceEngage,
		Kind:           KindEngageToolRun,
		IdempotencyKey: EngageToolRunIdempotencyKey("nuclei", "https://example.com", "2026-05-16T00:00:00Z"),
		Payload:        payload,
	}
	if err := env.Validate(); err != nil {
		t.Fatal(err)
	}
}
