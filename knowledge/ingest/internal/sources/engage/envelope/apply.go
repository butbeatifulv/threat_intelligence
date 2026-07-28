package ingest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/butbeautifulv/veil/pkg/commit"

	neo4jstore "github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/engage/storage"
)

// ApplyEnvelope applies engage kinds to Neo4j.
func ApplyEnvelope(ctx context.Context, st *neo4jstore.Store, env *commit.Envelope) error {
	switch env.Kind {
	case commit.KindEngageToolRun:
		var p commit.EngageToolRunPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.UpsertToolRun(ctx, env.IdempotencyKey, p.Tool, p.Target, p.Subject, p.Success, p.At)
	case commit.KindEngageFinding:
		var p commit.EngageFindingPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return st.UpsertFinding(ctx, env.IdempotencyKey, p.Tool, p.Target, p.Title, p.Severity, p.Description)
	default:
		return fmt.Errorf("engage graph ingest: unknown kind %q", env.Kind)
	}
}
