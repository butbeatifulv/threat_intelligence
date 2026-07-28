package mcp

import (
	"context"
	"errors"

	"github.com/butbeautifulv/veil/pkg/auth"
)

// AuthorizeToolCall enforces RBAC for tools/call when auth is enabled.
// fallback is used when no subject is in context (stdio / token path).
func AuthorizeToolCall(ctx context.Context, stack *auth.Stack, perm string, fallback func(context.Context, *auth.Stack, string) (context.Context, error)) (context.Context, error) {
	if stack == nil || !stack.Config.Enabled {
		return ctx, nil
	}
	if sub, ok := auth.SubjectFromContext(ctx); ok {
		if err := stack.Enforcer.Enforce(sub, perm); err != nil {
			if errors.Is(err, auth.ErrForbidden) {
				return ctx, Err(CodeAuthError, "forbidden")
			}
			return ctx, Err(CodeAuthError, "unauthorized")
		}
		return ctx, nil
	}
	if fallback == nil {
		return ctx, Err(CodeAuthError, "unauthorized")
	}
	ctx, err := fallback(ctx, stack, "")
	if err != nil {
		if errors.Is(err, auth.ErrForbidden) {
			return ctx, Err(CodeAuthError, "forbidden")
		}
		return ctx, Err(CodeAuthError, "unauthorized")
	}
	return ctx, nil
}
