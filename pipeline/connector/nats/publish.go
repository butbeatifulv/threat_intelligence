package nats

import (
	"context"
	"fmt"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/natsjet"
	natsgo "github.com/nats-io/nats.go"
)

// JetStreamPublisher publishes commit envelopes with deduplication header.
type JetStreamPublisher struct {
	conn *natsjet.Conn
}

func ConnectJetStream(url string) (*JetStreamPublisher, error) {
	conn, err := natsjet.Connect(url)
	if err != nil {
		return nil, err
	}
	return &JetStreamPublisher{conn: conn}, nil
}

func (p *JetStreamPublisher) Close() { p.conn.Close() }

// PublishJSON marshals the envelope and publishes to subject with Nats-Msg-Id.
func (p *JetStreamPublisher) PublishJSON(ctx context.Context, subject string, env *commit.Envelope) error {
	return natsjet.PublishCommitEnvelope(ctx, p.conn, subject, env)
}

// PublishCommit builds and publishes a commit envelope on subject.
func (p *JetStreamPublisher) PublishCommit(ctx context.Context, subject, source, kind, idempotencyKey string, payload any) error {
	return natsjet.PublishCommit(ctx, p.conn, subject, source, kind, idempotencyKey, payload)
}

// EnsureIngestStream creates or updates the INGEST stream to accept all ingest.* subjects.
func EnsureIngestStream(js natsgo.JetStreamContext) error {
	return natsjet.EnsureIngestStream(js)
}

// EnsureAppSecStream is kept for callers; it now ensures the unified ingest stream.
func EnsureAppSecStream(js natsgo.JetStreamContext) error {
	return natsjet.EnsureIngestStream(js)
}

// ConnectJetStreamAndStream connects and ensures INGEST stream exists.
func ConnectJetStreamAndStream(url string) (*JetStreamPublisher, error) {
	pub, err := ConnectJetStream(url)
	if err != nil {
		return nil, err
	}
	if err := EnsureAppSecStream(pub.conn.JS); err != nil {
		pub.Close()
		return nil, fmt.Errorf("ingest stream: %w", err)
	}
	return pub, nil
}

// EnsureBothStreams ensures SCRAPE and INGEST streams (for pipeline_worker).
func EnsureBothStreams(js natsgo.JetStreamContext) error {
	return natsjet.EnsureScrapeAndIngest(js)
}

// EnsureScrapeStream creates or updates SCRAPE (scrape.>).
func EnsureScrapeStream(js natsgo.JetStreamContext) error {
	return natsjet.EnsureScrapeStream(js)
}
