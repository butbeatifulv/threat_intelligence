package auth

import (
	"os"
	"strings"
)

// Config holds optional Keycloak JWT auth and RBAC settings.
type Config struct {
	Enabled        bool
	RBACEnabled    bool
	Issuer         string
	Audience       string
	ClientID       string
	RoleReader       string
	RoleAdmin        string
	RoleEngageRunner string
	RoleEngageAdmin  string
	MCPAccessToken      string
	StaticBearerToken   string // local pentest only; do not use in real prod
}

func LoadConfigFromEnv() Config {
	return Config{
		Enabled:        envBool("AUTH_ENABLED", false),
		RBACEnabled:    envBool("RBAC_ENABLED", false),
		Issuer:         strings.TrimSpace(os.Getenv("KEYCLOAK_ISSUER")),
		Audience:       strings.TrimSpace(os.Getenv("KEYCLOAK_AUDIENCE")),
		ClientID:       envOr("KEYCLOAK_CLIENT_ID", "veil-api"),
		RoleReader:       envOr("RBAC_ROLE_READER", "veil-reader"),
		RoleAdmin:        envOr("RBAC_ROLE_ADMIN", "veil-admin"),
		RoleEngageRunner: envOr("RBAC_ROLE_ENGAGE_RUNNER", "veil-engage-runner"),
		RoleEngageAdmin:  envOr("RBAC_ROLE_ENGAGE_ADMIN", "veil-engage-admin"),
		MCPAccessToken:    strings.TrimSpace(os.Getenv("MCP_ACCESS_TOKEN")),
		StaticBearerToken: strings.TrimSpace(os.Getenv("AUTH_STATIC_BEARER_TOKEN")),
	}
}

func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Issuer == "" && c.StaticBearerToken == "" {
		return ErrUnauthorized // wrapped at call site with better message
	}
	return nil
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

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
