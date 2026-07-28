package auth

import (
	"errors"
	"testing"
)

func TestLoadConfigFromEnv_defaults(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "")
	t.Setenv("RBAC_ENABLED", "")
	t.Setenv("KEYCLOAK_ISSUER", "")
	t.Setenv("KEYCLOAK_AUDIENCE", "")
	t.Setenv("KEYCLOAK_CLIENT_ID", "")
	t.Setenv("RBAC_ROLE_READER", "")
	t.Setenv("RBAC_ROLE_ADMIN", "")
	t.Setenv("RBAC_ROLE_ENGAGE_RUNNER", "")
	t.Setenv("RBAC_ROLE_ENGAGE_ADMIN", "")
	t.Setenv("MCP_ACCESS_TOKEN", "")
	t.Setenv("AUTH_STATIC_BEARER_TOKEN", "")

	cfg := LoadConfigFromEnv()
	if cfg.Enabled || cfg.RBACEnabled {
		t.Fatalf("expected auth/rbac disabled: %+v", cfg)
	}
	if cfg.ClientID != "veil-api" {
		t.Fatalf("ClientID: %q", cfg.ClientID)
	}
	if cfg.RoleReader != "veil-reader" || cfg.RoleAdmin != "veil-admin" {
		t.Fatalf("reader/admin roles: %+v", cfg)
	}
	if cfg.RoleEngageRunner != "veil-engage-runner" || cfg.RoleEngageAdmin != "veil-engage-admin" {
		t.Fatalf("engage roles: %+v", cfg)
	}
}

func TestLoadConfigFromEnv_envBool(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "false")
	t.Setenv("RBAC_ENABLED", "0")
	cfg := LoadConfigFromEnv()
	if cfg.Enabled || cfg.RBACEnabled {
		t.Fatalf("got %+v", cfg)
	}
}

func TestLoadConfigFromEnv_overrides(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("RBAC_ENABLED", "on")
	t.Setenv("KEYCLOAK_CLIENT_ID", "custom-client")
	t.Setenv("RBAC_ROLE_READER", "reader-x")

	cfg := LoadConfigFromEnv()
	if !cfg.Enabled || !cfg.RBACEnabled {
		t.Fatalf("enabled flags: %+v", cfg)
	}
	if cfg.ClientID != "custom-client" || cfg.RoleReader != "reader-x" {
		t.Fatalf("overrides: %+v", cfg)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr error
	}{
		{"disabled", Config{Enabled: false}, nil},
		{"enabled with issuer", Config{Enabled: true, Issuer: "https://kc/realms/v"}, nil},
		{"enabled with static token", Config{Enabled: true, StaticBearerToken: "secret"}, nil},
		{"enabled missing issuer and static", Config{Enabled: true}, ErrUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
