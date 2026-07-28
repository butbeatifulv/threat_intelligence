package config

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/pkg/auth"
)

// Config holds unified MCP gateway settings.
type Config struct {
	Listen         string
	Path           string
	GraphMCPURL    string
	EngageMCPURL   string
	Auth           auth.Config
	HTTPAuthStrict bool
	Prod           bool
}

func Load() Config {
	return Config{
		Listen:         envOr("UNIFIED_MCP_HTTP_LISTEN", ":8095"),
		Path:           envOr("UNIFIED_MCP_HTTP_PATH", "/mcp"),
		GraphMCPURL:    envOr("UNIFIED_MCP_GRAPH_URL", "http://127.0.0.1:8091/mcp"),
		EngageMCPURL:   envOr("UNIFIED_MCP_ENGAGE_URL", "http://127.0.0.1:8892/mcp"),
		Auth:           auth.LoadConfigFromEnv(),
		HTTPAuthStrict: envBool("UNIFIED_MCP_HTTP_AUTH_STRICT", false),
		Prod:           envBool("VEIL_PROD", false) || strings.EqualFold(strings.TrimSpace(os.Getenv("ENGAGE_ENV")), "prod"),
	}
}

func (c Config) HTTPClient() *http.Client {
	return &http.Client{Timeout: 120 * time.Second}
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
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
