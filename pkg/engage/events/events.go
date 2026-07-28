// Package events defines NATS wire types for engage.events.> (arXiv GAIA-adjacent bus).
package events

import "time"

// AuditEvent is published on engage.events.* when cross-layer NATS is enabled.
type AuditEvent struct {
	Source  string    `json:"source"`
	Tool    string    `json:"tool"`
	Target  string    `json:"target"`
	Subject string    `json:"subject"`
	Success bool      `json:"success"`
	At      time.Time `json:"at"`
}

// FindingEvent is published when smart-scan discovers vulnerabilities.
type FindingEvent struct {
	Tool        string `json:"tool"`
	Target      string `json:"target"`
	Title       string `json:"title"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}
