// Package domain holds shared Nuclei template entities for scrape, NED, and graph ingest.
package domain

// Template identifies a Nuclei YAML template path.
type Template struct {
	Repo string
	Path string
	ID   string
}
