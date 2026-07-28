package ingest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/butbeautifulv/veil/pkg/commit"

	neo4jstore "github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/ds/storage"
)

// ApplyEnvelope applies ds kinds to Neo4j.
func ApplyEnvelope(ctx context.Context, st *neo4jstore.Store, env *commit.Envelope) error {
	switch env.Kind {
	case commit.KindDSUpsertSigma:
		var p commit.DSUpsertSigmaPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.UpsertSigmaRule(ctx, p.ID, p.Title, p.Level, p.LogProduct, p.LogService, p.TagsJSON, p.Markdown, p.Source)
	case commit.KindDSUpsertYara:
		var p commit.DSUpsertYaraPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.UpsertYaraRule(ctx, p.ID, p.Name, p.Author, p.TagsJSON, p.Markdown, p.Source)
	case commit.KindDSUpsertAtomic:
		var p commit.DSUpsertAtomicPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.UpsertAtomicTest(ctx, p.ID, p.Name, p.Tactic, p.Technique, p.ExecName, p.ExecCmd, p.Markdown, p.Source)
	case commit.KindDSUpsertCaldera:
		var p commit.DSUpsertCalderaPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.UpsertCalderaAbility(ctx, p.ID, p.Name, p.Tactic, p.TechniqueID, p.Markdown, p.Source)
	default:
		return fmt.Errorf("ds graph ingest: unknown kind %q", env.Kind)
	}
}
