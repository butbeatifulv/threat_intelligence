package dedup

import (
	"context"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/transform"
)

// IngestPublisher publishes validated commit envelopes (JetStream or test double).
type IngestPublisher interface {
	PublishJSON(ctx context.Context, subject string, env *commit.Envelope) error
}

// PublishIngest publishes envelopes with JetStream dedup via Nats-Msg-Id (idempotency_key).
func PublishIngest(ctx context.Context, pub IngestPublisher, subject string, envs []*commit.Envelope) error {
	for _, e := range envs {
		if err := pub.PublishJSON(ctx, subject, e); err != nil {
			return err
		}
	}
	return nil
}

// ProcessScrapeMessage runs NED on a scrape envelope and publishes ingest results.
func ProcessScrapeMessage(ctx context.Context, pub IngestPublisher, ingestSubject string, env *harvest.Envelope) error {
	out, err := transform.ScrapeToIngest(ctx, env)
	if err != nil {
		return err
	}
	return PublishIngest(ctx, pub, ingestSubject, out)
}
