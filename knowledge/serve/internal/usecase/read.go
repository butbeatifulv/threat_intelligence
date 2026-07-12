package usecase

import (
	"context"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/butbeautifulv/veil/knowledge/connector/query"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/domain"
	"github.com/butbeautifulv/veil/pkg/retrieval"
)

// ReadUsecase exposes read-only graph operations for HTTP API and MCP.
type ReadUsecase struct {
	*query.Service
	exec        query.ReadExecutor
	categoryReg *retrieval.CategoryRegistry
}

func NewReadUsecase(exec query.ReadExecutor) *ReadUsecase {
	return &ReadUsecase{
		Service: query.NewService(exec),
		exec:    exec,
	}
}

func (u *ReadUsecase) ListCategoryMeta() []query.CategoryMeta {
	return query.ListCategoryMeta()
}

// GetNodeForAPI maps a missing node to domain.ErrNodeNotFound (HTTP 404).
func (u *ReadUsecase) GetNodeForAPI(ctx context.Context, id string) (*query.Node, error) {
	n, err := u.GetNode(ctx, id)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, domain.ErrNodeNotFound
	}
	return n, nil
}

// EngageTargetContext returns engage subgraph for a normalized host name.
func (u *ReadUsecase) EngageTargetContext(ctx context.Context, host string) (*query.EngageTargetContext, error) {
	return u.Service.EngageTargetContext(ctx, host)
}

// Ping verifies Neo4j read connectivity.
func (u *ReadUsecase) Ping(ctx context.Context) error {
	_, err := u.exec.ExecRead(ctx, func(tx driver.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, `RETURN 1`, nil)
		return nil, err
	})
	return err
}
