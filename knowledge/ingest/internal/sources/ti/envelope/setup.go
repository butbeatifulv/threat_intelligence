package envelope

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/butbeautifulv/veil/pkg/commit"

	neo4jstore "github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/ti/storage"
	"github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/ti/usecase"
)

// NeoConfig is Bolt credentials for the TI Neo4j writer.
type NeoConfig struct {
	URI, Username, Password, Database string
}

// SetupWriter opens the TI graph store and returns schema + apply + close for ingest_worker.
func SetupWriter(ctx context.Context, cfg NeoConfig, log *slog.Logger) (
	ensureSchema func(context.Context) error,
	apply func(context.Context, *commit.Envelope) error,
	closeFn func(context.Context) error,
	err error,
) {
	st, err := neo4jstore.New(ctx, neo4jstore.Config{
		URI:      cfg.URI,
		Username: cfg.Username,
		Password: cfg.Password,
		Database: cfg.Database,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("ti neo4j: %w", err)
	}
	uc := usecase.NewIngestor(st, log)
	ensureSchema = st.EnsureSchema
	apply = func(ctx context.Context, env *commit.Envelope) error {
		return ApplyEnvelope(ctx, st, uc, env)
	}
	closeFn = st.Close
	return ensureSchema, apply, closeFn, nil
}
