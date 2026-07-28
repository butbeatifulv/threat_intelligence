package mcpserver

import (
	"context"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// mockExec implements query.ReadExecutor for MCP transport tests (no real Neo4j).
type mockExec struct {
	failPing bool
}

func (m *mockExec) ExecRead(ctx context.Context, fn func(tx driver.ManagedTransaction) (any, error)) (any, error) {
	_ = fn
	if m.failPing {
		return nil, context.Canceled
	}
	return nil, nil
}
