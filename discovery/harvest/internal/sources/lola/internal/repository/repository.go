package repository

import (
	"context"

	"github.com/butbeautifulv/veil/pkg/lola/domain"
)

type LolaRepository interface {
	EnsureSchema(ctx context.Context) error
	UpsertArtifact(ctx context.Context, source string, a *domain.Artifact) error
	UpsertLoftsEntry(ctx context.Context, title, category, linkURL, markdown string) error
	UpsertAttackTechnique(ctx context.Context, id, name, description, markdown string) error
	UpsertAttackTactic(ctx context.Context, id, name, description, markdown string) error
	MergeTacticIncludesTechnique(ctx context.Context, tacticID, techniqueID string) error
	MergeSubtechniqueOf(ctx context.Context, parentTechniqueID, childTechniqueID string) error
	LinkArtifactsAndCommandsToTechniques(ctx context.Context) error
}
