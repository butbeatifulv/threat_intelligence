package neo4j

import (
	"context"

	"github.com/butbeautifulv/veil/knowledge/serve/internal/connector"
)

type Config = connector.Config

// Store implements query.ReadExecutor (shared by API and MCP).
type Store = connector.ReadConnector

// New opens a read-only Bolt connection.
func New(ctx context.Context, cfg Config) (*Store, error) {
	return connector.NewRead(ctx, cfg)
}
