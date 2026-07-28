package repository

import "context"

// GraphStore publishes raw scrape events (pipeline-worker → ingest.>).
type GraphStore interface {
	EnsureSchema(ctx context.Context) error
	UpsertSigmaRaw(ctx context.Context, path, rawYAML string) error
	UpsertYaraRaw(ctx context.Context, path, name, rawBody string) error
	UpsertAtomicRaw(ctx context.Context, path, rawYAML string) error
	UpsertCalderaRaw(ctx context.Context, path, fileName, rawBody string) error
}
