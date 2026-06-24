package nats

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/butbeautifulv/veil/pkg/commit"
	engageevents "github.com/butbeautifulv/veil/pkg/engage/events"
)

// Contract: engage.events.audit → commit engage_tool_run (veneno → veil bridge).
func TestContractEngageAuditRoundtrip(t *testing.T) {
	at := time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC)
	wire := engageevents.AuditEvent{
		Source: "veneno", Tool: "nmap", Target: "127.0.0.1",
		Subject: "scan", Success: true, At: at,
	}
	raw, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	var decoded engageevents.AuditEvent
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	atStr := decoded.At.UTC().Format(time.RFC3339)
	payload, err := json.Marshal(commit.EngageToolRunPayload{
		Tool: decoded.Tool, Target: decoded.Target, Subject: decoded.Subject,
		Success: decoded.Success, At: atStr,
	})
	if err != nil {
		t.Fatal(err)
	}
	env := commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceEngage,
		Kind:           commit.KindEngageToolRun,
		IdempotencyKey: commit.EngageToolRunIdempotencyKey(decoded.Tool, decoded.Target, atStr),
		Payload:        payload,
	}
	if err := env.Validate(); err != nil {
		t.Fatal(err)
	}
}

// Contract: engage.events.finding → commit engage_finding.
func TestContractEngageFindingRoundtrip(t *testing.T) {
	wire := engageevents.FindingEvent{
		Tool: "nuclei", Target: "https://example.com", Title: "xss",
		Severity: "high", Description: "reflected",
	}
	raw, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	var decoded engageevents.FindingEvent
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(commit.EngageFindingPayload{
		Tool: decoded.Tool, Target: decoded.Target, Title: decoded.Title,
		Severity: decoded.Severity, Description: decoded.Description,
	})
	if err != nil {
		t.Fatal(err)
	}
	env := commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceEngage,
		Kind:           commit.KindEngageFinding,
		IdempotencyKey: commit.EngageFindingIdempotencyKey(decoded.Tool, decoded.Target, decoded.Title),
		Payload:        payload,
	}
	if err := env.Validate(); err != nil {
		t.Fatal(err)
	}
}
