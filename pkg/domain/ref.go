package domain

import "strings"

// SourceRef is the shared thin-identity pattern for ingest entities (source + key + path + kind).
type SourceRef struct {
	Source Source
	Key    string
	Path   string
	Kind   string
}

// IsZero reports whether the ref has no identifying fields.
func (r SourceRef) IsZero() bool {
	return r.Source == "" && strings.TrimSpace(r.Key) == "" &&
		strings.TrimSpace(r.Path) == "" && strings.TrimSpace(r.Kind) == ""
}

// Valid reports whether the ref has a registered source and a non-empty key or path.
func (r SourceRef) Valid() bool {
	if !r.Source.Valid() {
		return false
	}
	return strings.TrimSpace(r.Key) != "" || strings.TrimSpace(r.Path) != ""
}
