package version

import "os"

const (
	ServerName = "veil-mcp"
	Default    = "0.4.2"
)

// MCP returns the MCP server version (APP_VERSION or MCP_VERSION env, else Default).
func MCP() string {
	if v := os.Getenv("MCP_VERSION"); v != "" {
		return v
	}
	if v := os.Getenv("APP_VERSION"); v != "" {
		return v
	}
	return Default
}
