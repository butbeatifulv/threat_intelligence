package repository

import (
	"context"
)

// CoderulesWriter is implemented by Neo4j storage for coderules ingest.
type CoderulesWriter interface {
	EnsureSchema(ctx context.Context) error
	Close(ctx context.Context) error
	UpsertCWECatalog(ctx context.Context, cweID, name, description, status string) error
	UpsertSemgrepRule(ctx context.Context, id, path, title, lang, markdown string) error
	LinkSemgrepRuleToCWE(ctx context.Context, ruleID, cweID string) error
	UpsertCodeQLRule(ctx context.Context, id, path, name, lang, markdown string) error
	LinkCodeQLRuleToCWE(ctx context.Context, ruleID, cweID string) error
}
