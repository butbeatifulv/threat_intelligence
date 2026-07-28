package nats

import (
	"context"
	"fmt"

	"github.com/butbeautifulv/veil/pkg/natsjet"
	"github.com/butbeautifulv/veil/pkg/harvest"
	natsgo "github.com/nats-io/nats.go"
)

// JetStreamPublisher publishes harvest envelopes to scrape.> subjects.
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

func (p *JetStreamPublisher) PublishJSON(ctx context.Context, subject string, env *harvest.Envelope) error {
	return natsjet.PublishHarvestEnvelope(ctx, p.conn, subject, env)
}

// PublishHarvest builds and publishes a harvest envelope on subject.
func (p *JetStreamPublisher) PublishHarvest(ctx context.Context, subject, source, kind, contentKey string, payload any) error {
	return natsjet.PublishHarvest(ctx, p.conn, subject, source, kind, contentKey, payload)
}

// EnsureScrapeStream creates or updates the SCRAPE stream (scrape.>).
func EnsureScrapeStream(js natsgo.JetStreamContext) error {
	return natsjet.EnsureScrapeStream(js)
}

// ConnectJetStreamAndStream connects and ensures SCRAPE stream exists.
func ConnectJetStreamAndStream(url string) (*JetStreamPublisher, error) {
	pub, err := ConnectJetStream(url)
	if err != nil {
		return nil, err
	}
	if err := EnsureScrapeStream(pub.conn.JS); err != nil {
		pub.Close()
		return nil, fmt.Errorf("scrape stream: %w", err)
	}
	return pub, nil
}
