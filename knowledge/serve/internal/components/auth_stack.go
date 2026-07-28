package components

import (
	"context"
	"fmt"

	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/auth/keycloak"
	"github.com/butbeautifulv/veil/pkg/auth/static"
)

func newAuthStack(ctx context.Context, cfg auth.Config) (*auth.Stack, error) {
	if !cfg.Enabled {
		return auth.NewStack(nil, cfg), nil
	}
	if tok := cfg.StaticBearerToken; tok != "" {
		return auth.NewStack(static.New(tok, "pentest-runner", nil), cfg), nil
	}
	if cfg.Issuer == "" {
		return nil, fmt.Errorf("AUTH_ENABLED=1 requires KEYCLOAK_ISSUER or AUTH_STATIC_BEARER_TOKEN")
	}
	v, err := keycloak.NewVerifier(ctx, cfg.Issuer, cfg.Audience, cfg.ClientID)
	if err != nil {
		return nil, err
	}
	return auth.NewStack(v, cfg), nil
}
