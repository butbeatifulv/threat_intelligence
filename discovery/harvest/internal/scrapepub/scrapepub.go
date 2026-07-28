// Package scrapepub provides shared scrape→NATS publish wiring for harvest sources.
package scrapepub

import (
	"context"

	connats "github.com/butbeautifulv/veil/discovery/connector/nats"
)

// RawPublisher publishes harvest for one domain source and subject.
type RawPublisher interface {
	Publish(ctx context.Context, kind, contentKey string, payload any) error
}

// Base holds the raw publisher used by per-source scrapepub adapters.
type Base struct {
	Raw RawPublisher
}

// NewBase wraps a raw publisher for embedding in source-specific types.
func NewBase(raw RawPublisher) Base {
	return Base{Raw: raw}
}

// NewRaw wires JetStream to a harvest domain publisher for source on subject.
func NewRaw(pub *connats.JetStreamPublisher, source, subject string) RawPublisher {
	return connats.NewDomainPublisher(pub, source, subject)
}
