package api

import (
	"net/http"

	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/auth/httpmiddleware"
)

// AuthMiddleware wraps handlers with JWT Bearer auth and RBAC for perm.
// When strict is true, only GET /health is public.
func AuthMiddleware(stack *auth.Stack, strict, prod bool, perm string, next http.Handler) http.Handler {
	return httpmiddleware.Auth(stack, strict, prod, perm, next)
}
