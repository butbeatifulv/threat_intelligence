package config

import (
	"os"
	"strings"
)

// MCPHTTPConfig holds Streamable HTTP transport settings for veil-mcp.
type MCPHTTPConfig struct {
	Enabled   bool
	Listen    string
	Path      string
	PreferSSE bool
	BindLocal bool
}

func loadMCPHTTPFromEnv() MCPHTTPConfig {
	return MCPHTTPConfig{
		Enabled:   envBool("MCP_HTTP_ENABLED", false),
		Listen:    getenv("MCP_HTTP_LISTEN", ":8091"),
		Path:      getenv("MCP_HTTP_PATH", "/mcp"),
		PreferSSE: envBool("MCP_HTTP_PREFER_SSE", false),
		BindLocal: envBool("MCP_HTTP_BIND_LOCAL", false),
	}
}

// ResolveListen returns the listen address, optionally binding to localhost only.
func (c MCPHTTPConfig) ResolveListen() string {
	addr := strings.TrimSpace(c.Listen)
	if addr == "" {
		addr = ":8091"
	}
	if c.BindLocal && strings.HasPrefix(addr, ":") {
		return "127.0.0.1" + addr
	}
	return addr
}

func envBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
