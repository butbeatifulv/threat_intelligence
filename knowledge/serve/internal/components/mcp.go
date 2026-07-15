package components

import (
	"context"
	"log/slog"
	"net/http"

	authmw "github.com/butbeautifulv/veil/knowledge/serve/internal/auth/middleware"
	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/config"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/connector"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/transport/mcpserver"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/transport/securityhttp"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/usecase"
	playbookuc "github.com/butbeautifulv/veil/knowledge/serve/internal/usecase/playbook"
	procedureuc "github.com/butbeautifulv/veil/knowledge/serve/internal/usecase/procedure"
)

type MCPComponents struct {
	Neo4j     *connector.ReadConnector
	Read      *usecase.ReadUsecase
	Playbook  *playbookuc.Service
	Procedure *procedureuc.Service
	Auth      *auth.Stack
	MCPServer *mcpserver.Server
}

func InitMCP(cfg *config.Config, logger *slog.Logger) (*MCPComponents, error) {
	conn, err := connector.NewRead(context.Background(), connector.Config{
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
		_ = conn.Close(context.Background())
		return nil, err
	}
	uc := usecase.NewReadUsecase(conn)
	pb, err := playbookuc.NewService()
	if err != nil {
		_ = conn.Close(context.Background())
		return nil, err
	}
	proc, err := procedureuc.NewService()
	if err != nil {
		_ = conn.Close(context.Background())
		return nil, err
	}
	return &MCPComponents{
		Neo4j:     conn,
		Read:      uc,
		Playbook:  pb,
		Procedure: proc,
		Auth:      stack,
		MCPServer: mcpserver.NewServer(uc, pb, proc, nil, stack, logger),
	}, nil
}

// MCPHTTPHandler returns the Streamable HTTP handler with security and optional JWT middleware.
// When MCP_HTTP_AUTH_STRICT=1, every route except GET /health requires Bearer auth.
// Otherwise tool calls are authorized inside ProcessMessage (initialize may be unauthenticated).
func (c *MCPComponents) MCPHTTPHandler(cfg *config.Config) http.Handler {
	h := mcpserver.HTTPHandler(c.MCPServer, cfg.MCPHTTP)
	inner := h
	if c.Auth != nil && c.Auth.Config.Enabled && cfg.Security.MCPHTTPAuthStrict {
		inner = authmw.Auth(c.Auth, true, cfg.Security, h)
	}
	return securityhttp.Harden(cfg.Security, cfg.Security.MCPBodyLimit, inner)
}

func (c *MCPComponents) Shutdown() {
	if c.Neo4j != nil {
		_ = c.Neo4j.Close(context.Background())
	}
}
