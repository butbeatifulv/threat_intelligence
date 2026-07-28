package components

import (
	"context"

	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/config"
	neo4jstore "github.com/butbeautifulv/veil/knowledge/serve/internal/storage/neo4j"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/usecase"
	playbookuc "github.com/butbeautifulv/veil/knowledge/serve/internal/usecase/playbook"
	frameworkuc "github.com/butbeautifulv/veil/knowledge/serve/internal/usecase/framework"
	procedureuc "github.com/butbeautifulv/veil/knowledge/serve/internal/usecase/procedure"
)

type APIComponents struct {
	Neo4jStore *neo4jstore.Store
	Read       *usecase.ReadUsecase
	Playbook   *playbookuc.Service
	Procedure  *procedureuc.Service
	Framework  *frameworkuc.Service
	Auth       *auth.Stack
}

func InitAPI(cfg *config.Config) (*APIComponents, error) {
	if err := config.ValidateSecurity(cfg.Security, cfg.Auth.Enabled); err != nil {
		return nil, err
	}
	store, err := neo4jstore.New(context.Background(), neo4jstore.Config{
		URI:      cfg.Neo4j.URI,
		Username: cfg.Neo4j.Username,
		Password: cfg.Neo4j.Password,
		Database: cfg.Neo4j.Database,
	})
	if err != nil {
		return nil, err
	}
	stack, err := newAuthStack(context.Background(), cfg.Auth)
	if err != nil {
		_ = store.Close(context.Background())
		return nil, err
	}
	pb, err := playbookuc.NewService()
	if err != nil {
		_ = store.Close(context.Background())
		return nil, err
	}
	proc, err := procedureuc.NewService()
	if err != nil {
		_ = store.Close(context.Background())
		return nil, err
	}
	return &APIComponents{
		Neo4jStore: store,
		Read:       usecase.NewReadUsecase(store),
		Playbook:   pb,
		Procedure:  proc,
		Framework:  frameworkuc.NewService(),
		Auth:       stack,
	}, nil
}

func (c *APIComponents) Shutdown() {
	if c.Neo4jStore != nil {
		_ = c.Neo4jStore.Close(context.Background())
	}
}
