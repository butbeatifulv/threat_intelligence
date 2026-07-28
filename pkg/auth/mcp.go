package auth

import "context"

// AuthorizeMCP validates the MCP access token and enforces graph read permission.
func AuthorizeMCP(ctx context.Context, stack *Stack, rawToken string) (context.Context, error) {
	if stack == nil || !stack.Config.Enabled {
		return ctx, nil
	}
	if sub, ok := SubjectFromContext(ctx); ok {
		if err := stack.Enforcer.Enforce(sub, PermGraphRead); err != nil {
			return ctx, err
		}
		return ctx, nil
	}
	if stack.Verifier == nil {
		return ctx, ErrUnauthorized
	}
	raw := rawToken
	if raw == "" {
		raw = stack.Config.MCPAccessToken
	}
	if raw == "" {
		return ctx, ErrUnauthorized
	}
	sub, err := stack.Verifier.Validate(ctx, raw)
	if err != nil {
		return ctx, err
	}
	if err := stack.Enforcer.Enforce(sub, PermGraphRead); err != nil {
		return ctx, err
	}
	return WithSubject(ctx, sub), nil
}

// AuthorizeEngageMCP validates the MCP access token and enforces engage tool run permission.
func AuthorizeEngageMCP(ctx context.Context, stack *Stack, rawToken string) (context.Context, error) {
	if stack == nil || !stack.Config.Enabled {
		return ctx, nil
	}
	if sub, ok := SubjectFromContext(ctx); ok {
		if err := stack.Enforcer.Enforce(sub, PermEngageToolRun); err != nil {
			return ctx, err
		}
		return ctx, nil
	}
	if stack.Verifier == nil {
		return ctx, ErrUnauthorized
	}
	raw := rawToken
	if raw == "" {
		raw = stack.Config.MCPAccessToken
	}
	if raw == "" {
		return ctx, ErrUnauthorized
	}
	sub, err := stack.Verifier.Validate(ctx, raw)
	if err != nil {
		return ctx, err
	}
	if err := stack.Enforcer.Enforce(sub, PermEngageToolRun); err != nil {
		return ctx, err
	}
	return WithSubject(ctx, sub), nil
}
