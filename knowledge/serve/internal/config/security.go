package config

import (
	"errors"
	"os"
	"strings"
)

// SecurityConfig holds deployment hardening flags.
type SecurityConfig struct {
	RequireAuth       bool
	MCPHTTPAuthStrict bool
	Prod              bool
	CORSAllowedOrigins []string
	APIBodyLimit      int64
	MCPBodyLimit      int64
}

// LoadSecurityForEnv builds security settings for the given environment name (e.g. prod).
func LoadSecurityForEnv(env string) SecurityConfig {
	return loadSecurityFromEnv(env)
}

func loadSecurityFromEnv(env string) SecurityConfig {
	prod := strings.EqualFold(strings.TrimSpace(env), "prod")
	cors := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	var origins []string
	if cors != "" {
		for _, o := range strings.Split(cors, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				origins = append(origins, o)
			}
		}
	}
	return SecurityConfig{
		RequireAuth:       envBool("VEIL_REQUIRE_AUTH", false),
		MCPHTTPAuthStrict: envBool("MCP_HTTP_AUTH_STRICT", false),
		Prod:              prod,
		CORSAllowedOrigins: origins,
		APIBodyLimit:      1 << 20,
		MCPBodyLimit:      4 << 20,
	}
}

// ValidateSecurity fails closed when production auth is required but disabled.
func ValidateSecurity(sec SecurityConfig, authEnabled bool) error {
	if sec.RequireAuth && !authEnabled {
		return errors.New("VEIL_REQUIRE_AUTH=1 requires AUTH_ENABLED=1")
	}
	return nil
}
