package events

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAuditEvent_JSONRoundTrip(t *testing.T) {
	at := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	in := AuditEvent{
		Source: "veil-engage", Tool: "nmap", Target: "example.com",
		Subject: "engage.events.audit", Success: true, At: at,
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out AuditEvent
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.Tool != in.Tool || !out.At.Equal(at) {
		t.Fatalf("got %+v", out)
	}
}

func TestFindingEvent_JSONRoundTrip(t *testing.T) {
	in := FindingEvent{Tool: "nuclei", Target: "https://x", Title: "xss", Severity: "high"}
	b, _ := json.Marshal(in)
	var out FindingEvent
	if err := json.Unmarshal(b, &out); err != nil || out.Title != "xss" {
		t.Fatalf("out=%+v", out)
	}
}
