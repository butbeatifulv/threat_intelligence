// Package domain holds shared detections/signals entities for scrape, NED, and graph ingest.
package domain

// Resource identifies a detections/signals artifact (Caldera, Sigma, etc.).
type Resource struct {
	Key    string
	Source string
	URL    string
	Kind   string
}
