package repository

import "context"

// NucleiWriter is implemented by Neo4j storage for nuclei ingest.
type NucleiWriter interface {
	EnsureSchema(ctx context.Context) error
	Close(ctx context.Context) error
	UpsertNucleiTemplate(ctx context.Context, id, templateKey, path, name, severity, tagsJSON, cveID, cweID, markdown string) error
}
