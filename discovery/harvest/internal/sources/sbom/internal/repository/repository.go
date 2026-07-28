package repository

import (
	"context"
)

// SBOMWriter is implemented by Neo4j storage for SBOM ingest.
type SBOMWriter interface {
	EnsureSchema(ctx context.Context) error
	Close(ctx context.Context) error
	ListCVEs(ctx context.Context, limit int) ([]string, error)
	UpsertFromOSVVuln(ctx context.Context, osvID string, cve string, affected []map[string]any) error
	UpsertGHSA(ctx context.Context, doc map[string]any) error
}
