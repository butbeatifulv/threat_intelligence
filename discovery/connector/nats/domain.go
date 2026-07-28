package nats

import (
	"context"
	"strings"
)

// DomainPublisher publishes harvest envelopes for one domain source and subject.
type DomainPublisher struct {
	Source  string
	Pub     *JetStreamPublisher
	Subject string
}

// NewDomainPublisher returns a publisher for source on subject.
func NewDomainPublisher(pub *JetStreamPublisher, source string, subject string) *DomainPublisher {
	return &DomainPublisher{
		Source:  source,
		Pub:     pub,
		Subject: strings.TrimSpace(subject),
	}
}

// Publish builds a harvest envelope and publishes to JetStream.
func (p *DomainPublisher) Publish(ctx context.Context, kind, contentKey string, payload any) error {
	return p.Pub.PublishHarvest(ctx, p.Subject, p.Source, kind, contentKey, payload)
}
