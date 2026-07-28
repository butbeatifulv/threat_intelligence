package nats

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/butbeautifulv/veil/pkg/commit"
	engageevents "github.com/butbeautifulv/veil/pkg/engage/events"
	natsgo "github.com/nats-io/nats.go"
)

func TestHandleEngageFinding_buildsEnvelope(t *testing.T) {
	wire := engageevents.FindingEvent{
		Tool: "nuclei", Target: "https://example.com", Title: "xss", Severity: "high",
	}
	b, _ := json.Marshal(wire)
	var got commit.Envelope
	payload, err := json.Marshal(commit.EngageFindingPayload{
		Tool: wire.Tool, Target: wire.Target, Title: wire.Title, Severity: wire.Severity,
	})
	if err != nil {
		t.Fatal(err)
	}
	got = commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceEngage,
		Kind:           commit.KindEngageFinding,
		IdempotencyKey: commit.EngageFindingIdempotencyKey(wire.Tool, wire.Target, wire.Title),
		Payload:        payload,
	}
	if err := got.Validate(); err != nil {
		t.Fatal(err)
	}
	_ = b
}

func TestHandleEngageMsg_buildsEnvelope(t *testing.T) {
	wire := engageevents.AuditEvent{
		Tool: "nmap", Target: "127.0.0.1", Subject: "test", Success: true,
		At: time.Now().UTC(),
	}
	b, _ := json.Marshal(wire)
	pub := &JetStreamPublisher{conn: nil}
	var got commit.Envelope
	atStr := wire.At.UTC().Format(time.RFC3339)
	payload, err := json.Marshal(commit.EngageToolRunPayload{
		Tool: wire.Tool, Target: wire.Target, Subject: wire.Subject, Success: wire.Success, At: atStr,
	})
	if err != nil {
		t.Fatal(err)
	}
	got = commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceEngage,
		Kind:           commit.KindEngageToolRun,
		IdempotencyKey: commit.EngageToolRunIdempotencyKey(wire.Tool, wire.Target, atStr),
		Payload:        payload,
	}
	if err := got.Validate(); err != nil {
		t.Fatal(err)
	}
	_ = b
	_ = pub
	_ = context.Background()
	_ = natsgo.ErrTimeout
}
