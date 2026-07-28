package mcpserver

import (
	"log/slog"
	"net/http"

	"github.com/butbeautifulv/veil/knowledge/serve/internal/config"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/version"
	"github.com/butbeautifulv/veil/pkg/mcp"
)

// HTTPHandler serves Streamable HTTP MCP (POST JSON or SSE).
func HTTPHandler(s *Server, cfg config.MCPHTTPConfig) http.Handler {
	path := cfg.Path
	if path == "" {
		path = "/mcp"
	}
	return mcp.HTTPHandler(s, mcp.HTTPConfig{
		Path:      path,
		PreferSSE: cfg.PreferSSE,
		Service:   version.ServerName,
		HealthExtra: map[string]any{
			"transport": "streamable-http",
		},
		Logger: slog.Default(),
	})
}
