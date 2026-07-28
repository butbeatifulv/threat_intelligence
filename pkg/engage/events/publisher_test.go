package events

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

func startTestNATS(t *testing.T) string {
	t.Helper()
	opts := &server.Options{JetStream: true, StoreDir: t.TempDir(), Port: -1}
	srv := test.RunServer(opts)
	t.Cleanup(srv.Shutdown)
	return srv.ClientURL()
}

func TestPublisher_nilNoOp(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var p *Publisher
	if err := p.PublishAudit(ctx, AuditEvent{Tool: "nmap"}); err != nil {
		t.Fatal(err)
	}
	if err := p.PublishFinding(ctx, FindingEvent{Title: "x"}); err != nil {
		t.Fatal(err)
	}
	p = &Publisher{}
	if err := p.PublishAudit(ctx, AuditEvent{Tool: "nmap"}); err != nil {
		t.Fatal(err)
	}
}

func TestConnectWithSubjects_publishAuditAndFinding(t *testing.T) {
	url := startTestNATS(t)
	auditSub := "engage.events.audit.test"
	findingSub := "engage.events.finding.test"

	pub, err := ConnectWithSubjects(url, auditSub, findingSub)
	if err != nil {
		t.Fatal(err)
	}
	defer pub.Close()

	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()
	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	at := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	if err := pub.PublishAudit(ctx, AuditEvent{
		Tool: "nmap", Target: "example.com", Subject: auditSub, Success: true, At: at,
	}); err != nil {
		t.Fatal(err)
	}
	if err := pub.PublishFinding(ctx, FindingEvent{
		Tool: "nuclei", Target: "https://example.com", Title: "xss", Severity: "high",
	}); err != nil {
		t.Fatal(err)
	}

	for _, subj := range []string{auditSub, findingSub} {
		msg, err := js.GetLastMsg("ENGAGE_EVENTS", subj)
		if err != nil {
			t.Fatalf("GetLastMsg %s: %v", subj, err)
		}
		if len(msg.Data) == 0 {
			t.Fatalf("empty payload on %s", subj)
		}
	}
}

func TestPublishAudit_setsSourceAndAt(t *testing.T) {
	url := startTestNATS(t)
	auditSub := "engage.events.audit.stamp"

	pub, err := ConnectWithSubjects(url, auditSub, "")
	if err != nil {
		t.Fatal(err)
	}
	defer pub.Close()

	if err := pub.PublishAudit(context.Background(), AuditEvent{
		Tool: "httpx", Target: "host", Success: false,
	}); err != nil {
		t.Fatal(err)
	}

	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()
	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	msg, err := js.GetLastMsg("ENGAGE_EVENTS", auditSub)
	if err != nil {
		t.Fatal(err)
	}
	var out AuditEvent
	if err := json.Unmarshal(msg.Data, &out); err != nil {
		t.Fatal(err)
	}
	if out.Source != "veil-engage" {
		t.Fatalf("source %q", out.Source)
	}
	if out.At.IsZero() {
		t.Fatal("expected non-zero At")
	}
}

func TestConnectWithSubjects_noJetStream(t *testing.T) {
	url := startTestNATSNoJS(t)
	_, err := ConnectWithSubjects(url, "engage.events.audit", "engage.events.finding")
	if err == nil {
		t.Fatal("expected stream ensure error")
	}
}

func startTestNATSNoJS(t *testing.T) string {
	t.Helper()
	opts := &server.Options{JetStream: false, Port: -1}
	srv := test.RunServer(opts)
	t.Cleanup(srv.Shutdown)
	return srv.ClientURL()
}

func TestConnectWithSubjects_invalidURL(t *testing.T) {
	_, err := ConnectWithSubjects("nats://127.0.0.1:1", "a", "b")
	if err == nil {
		t.Fatal("expected connect error")
	}
}

func TestConnect_defaultFindingSubject(t *testing.T) {
	url := startTestNATS(t)
	pub, err := Connect(url, "engage.events.audit.default")
	if err != nil {
		t.Fatal(err)
	}
	pub.Close()
}
