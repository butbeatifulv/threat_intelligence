package transport

import (
	"net/http"
	"strings"

	"github.com/butbeautifulv/veil/platform/mcp-gateway/internal/aggregator"
	"github.com/butbeautifulv/veil/platform/mcp-gateway/internal/config"
	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/auth/httpmiddleware"
	"github.com/butbeautifulv/veil/pkg/mcp"
)

// Handler serves unified Streamable HTTP MCP and forwards Bearer tokens to backends.
func Handler(agg *aggregator.Aggregator, cfg config.Config, stack *auth.Stack) http.Handler {
	path := cfg.Path
	if path == "" {
		path = "/mcp"
	}
	base := mcp.HTTPHandler(agg, mcp.HTTPConfig{
		Path:    path,
		Service: aggregator.ServerName,
		HealthExtra: map[string]any{
			"backends": []string{"graph", "engage"},
		},
		Logger: agg.Logger(),
	})
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if tok := bearerToken(r.Header.Get("Authorization")); tok != "" {
			ctx = aggregator.WithBearerToken(ctx, tok)
		}
		base.ServeHTTP(w, r.WithContext(ctx))
	})
	if stack != nil && stack.Config.Enabled && cfg.HTTPAuthStrict {
		// Strict mode: require graph read for any MCP route except GET /health.
		// Individual tools/call still enforces engage vs graph inside the aggregator.
		return httpmiddleware.Auth(stack, true, cfg.Prod, auth.PermGraphRead, inner)
	}
	return inner
}

func bearerToken(h string) string {
	h = strings.TrimSpace(h)
	const prefix = "Bearer "
	if strings.HasPrefix(h, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(h, prefix))
	}
	return ""
}
