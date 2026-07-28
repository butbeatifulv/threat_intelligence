package aggregator

import "context"

type bearerKey struct{}

// WithBearerToken stores the raw Bearer token for upstream MCP backends.
func WithBearerToken(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return context.WithValue(ctx, bearerKey{}, token)
}

func bearerToken(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(bearerKey{}).(string)
	return v
}

func authorizationHeader(ctx context.Context) string {
	if tok := bearerToken(ctx); tok != "" {
		return "Bearer " + tok
	}
	return ""
}
