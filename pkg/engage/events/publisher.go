package events

import (
	"context"
	"time"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/natsjet"
)

// Publisher sends engage events to JetStream (ENGAGE_EVENTS).
type Publisher struct {
	conn           *natsjet.Conn
	auditSubject   string
	findingSubject string
}

// Connect opens NATS and ensures ENGAGE_EVENTS stream.
func Connect(url, auditSubject string) (*Publisher, error) {
	return ConnectWithSubjects(url, auditSubject, "engage.events.finding")
}

// ConnectWithSubjects configures audit and finding subjects.
func ConnectWithSubjects(url, auditSubject, findingSubject string) (*Publisher, error) {
	conn, err := natsjet.Connect(url)
	if err != nil {
		return nil, err
	}
	if err := natsjet.EnsureEngageEventsStream(conn.JS); err != nil {
		conn.Close()
		return nil, err
	}
	if findingSubject == "" {
		findingSubject = "engage.events.finding"
	}
	return &Publisher{conn: conn, auditSubject: auditSubject, findingSubject: findingSubject}, nil
}

func (p *Publisher) Close() {
	if p != nil && p.conn != nil {
		p.conn.Close()
	}
}

func (p *Publisher) PublishAudit(ctx context.Context, e AuditEvent) error {
	if p == nil || p.conn == nil {
		return nil
	}
	e.Source = "veil-engage"
	at := e.At.UTC()
	if at.IsZero() {
		at = time.Now().UTC()
		e.At = at
	}
	atStr := at.Format(time.RFC3339)
	msgID := commit.EngageToolRunIdempotencyKey(e.Tool, e.Target, atStr)
	return p.conn.PublishJSON(ctx, p.auditSubject, e, msgID)
}

func (p *Publisher) PublishFinding(ctx context.Context, e FindingEvent) error {
	if p == nil || p.conn == nil {
		return nil
	}
	msgID := commit.EngageFindingIdempotencyKey(e.Tool, e.Target, e.Title)
	return p.conn.PublishJSON(ctx, p.findingSubject, e, msgID)
}
