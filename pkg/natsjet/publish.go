package natsjet

import (
	"context"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
)

// PublishHarvestEnvelope validates env and publishes with ContentKey as Nats-Msg-Id.
func PublishHarvestEnvelope(ctx context.Context, c *Conn, subject string, env *harvest.Envelope) error {
	if err := env.Validate(); err != nil {
		return err
	}
	return c.PublishJSON(ctx, subject, env, env.ContentKey)
}

// PublishHarvest builds a harvest envelope and publishes it to subject.
func PublishHarvest(ctx context.Context, c *Conn, subject, source, kind, contentKey string, payload any) error {
	env, err := harvest.NewEnvelope(source, kind, contentKey, payload)
	if err != nil {
		return err
	}
	return PublishHarvestEnvelope(ctx, c, subject, env)
}

// PublishCommitEnvelope validates env and publishes with IdempotencyKey as Nats-Msg-Id.
func PublishCommitEnvelope(ctx context.Context, c *Conn, subject string, env *commit.Envelope) error {
	if err := env.Validate(); err != nil {
		return err
	}
	return c.PublishJSON(ctx, subject, env, env.IdempotencyKey)
}

// PublishCommit builds a commit envelope and publishes it to subject.
func PublishCommit(ctx context.Context, c *Conn, subject, source, kind, idempotencyKey string, payload any) error {
	env, err := commit.NewEnvelope(source, kind, idempotencyKey, payload)
	if err != nil {
		return err
	}
	return PublishCommitEnvelope(ctx, c, subject, env)
}
