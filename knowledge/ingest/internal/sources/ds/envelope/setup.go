package ingest

import (
	"context"

	"github.com/butbeautifulv/veil/knowledge/ingest/internal/ingestkit"
	"github.com/butbeautifulv/veil/pkg/commit"

	neo4jstore "github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/ds/storage"
)

// NeoConfig is Bolt credentials.
type NeoConfig = ingestkit.NeoConfig

// SetupWriter returns schema + apply + close for ingest_worker.
func SetupWriter(ctx context.Context, cfg NeoConfig) (
	ensureSchema func(context.Context) error,
	apply func(context.Context, *commit.Envelope) error,
	closeFn func(context.Context) error,
	err error,
) {
	return ingestkit.SetupWriter(ctx, func(ctx context.Context, c ingestkit.NeoConfig) (*neo4jstore.Store, error) {
		return neo4jstore.New(ctx, neo4jstore.Config{
			URI: c.URI, Username: c.Username, Password: c.Password, Database: c.Database,
		})
	}, cfg, func(ctx context.Context, st *neo4jstore.Store, env *commit.Envelope) error {
		return ApplyEnvelope(ctx, st, env)
	})
}
