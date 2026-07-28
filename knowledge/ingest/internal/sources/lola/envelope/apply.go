package ingest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/butbeautifulv/veil/pkg/commit"

	"github.com/butbeautifulv/veil/pkg/lola/domain"
	neo4jstore "github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/lola/storage"
)

// ApplyEnvelope applies lola kinds to Neo4j.
func ApplyEnvelope(ctx context.Context, st *neo4jstore.Store, env *commit.Envelope) error {
	switch env.Kind {
	case commit.KindLolaArtifact:
		var p commit.LolaArtifactPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		var a domain.Artifact
		if err := json.Unmarshal(p.Body, &a); err != nil {
			return err
		}
		return st.UpsertArtifact(ctx, p.Source, &a)
	case commit.KindLolaLofts:
		var p commit.LolaLoftsPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.UpsertLoftsEntry(ctx, p.Title, p.Category, p.LinkURL, p.Markdown)
	case commit.KindLolaAttackTechnique:
		var p commit.LolaAttackTechniquePayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.UpsertAttackTechnique(ctx, p.ID, p.Name, p.Description, p.Markdown)
	case commit.KindLolaAttackTactic:
		var p commit.LolaAttackTacticPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.UpsertAttackTactic(ctx, p.ID, p.Name, p.Description, p.Markdown)
	case commit.KindLolaMergeTacticTechnique:
		var p commit.LolaMergeTacticTechniquePayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.MergeTacticIncludesTechnique(ctx, p.TacticID, p.TechniqueID)
	case commit.KindLolaMergeSubtechnique:
		var p commit.LolaMergeSubtechniquePayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.MergeSubtechniqueOf(ctx, p.ParentTechniqueID, p.ChildTechniqueID)
	case commit.KindLolaLinkArtifacts:
		return st.LinkArtifactsAndCommandsToTechniques(ctx)
	default:
		return fmt.Errorf("lola graph ingest: unknown kind %q", env.Kind)
	}
}
