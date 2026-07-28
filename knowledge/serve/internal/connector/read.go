package connector

import (
	"context"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"

	graphneo4j "github.com/butbeautifulv/veil/knowledge/connector/neo4j"
)

type Config = graphneo4j.Config

// ReadConnector implements query.ReadExecutor for MCP and API read paths.
type ReadConnector struct {
	client *graphneo4j.Client
}

func NewRead(ctx context.Context, cfg Config) (*ReadConnector, error) {
	c, err := graphneo4j.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &ReadConnector{client: c}, nil
}

func (c *ReadConnector) Close(ctx context.Context) error { return c.client.Close(ctx) }

func (c *ReadConnector) ExecRead(ctx context.Context, fn func(tx driver.ManagedTransaction) (any, error)) (any, error) {
	return c.client.ExecRead(ctx, fn)
}
