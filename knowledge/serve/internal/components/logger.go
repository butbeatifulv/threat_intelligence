package components

import (
	"log/slog"
	"os"

	"github.com/butbeautifulv/veil/pkg/observability"
)

func SetupLogger(env string) *slog.Logger {
	return observability.NewLogger(env, "veil-api", os.Stdout)
}

// SetupMCPLogger logs to stderr so stdio MCP JSON-RPC on stdout stays clean.
func SetupMCPLogger(env string) *slog.Logger {
	return observability.NewLogger(env, "veil-mcp", os.Stderr)
}
