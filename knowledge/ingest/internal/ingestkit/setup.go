// Package ingestkit provides shared ingest_worker wiring for graph domain stores.
package ingestkit

import (
	"context"

	"github.com/butbeautifulv/veil/pkg/commit"
)

// NeoConfig is Bolt credentials shared by domain ingest packages.
type NeoConfig struct {
	URI, Username, Password, Database string
}

// SchemaStore is the subset of Neo4j stores needed for ingest_worker setup.
type SchemaStore interface {
	EnsureSchema(context.Context) error
	Close(context.Context) error
}

// SetupWriter opens a store and returns schema + apply + close for ingest_worker.
func SetupWriter[S SchemaStore](
	ctx context.Context,
	newStore func(context.Context, NeoConfig) (S, error),
	cfg NeoConfig,
	applyEnvelope func(context.Context, S, *commit.Envelope) error,
) (
	ensureSchema func(context.Context) error,
	apply func(context.Context, *commit.Envelope) error,
	closeFn func(context.Context) error,
	err error,
) {
	st, err := newStore(ctx, cfg)
	if err != nil {
		return nil, nil, nil, err
	}
	ensureSchema = st.EnsureSchema
	apply = func(ctx context.Context, env *commit.Envelope) error {
		return applyEnvelope(ctx, st, env)
	}
	closeFn = st.Close
	return ensureSchema, apply, closeFn, nil
}
