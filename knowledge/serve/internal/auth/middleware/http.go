package middleware

import (
	"net/http"

	"github.com/butbeautifulv/veil/knowledge/serve/internal/config"
	"github.com/butbeautifulv/veil/pkg/api"
	"github.com/butbeautifulv/veil/pkg/auth"
)

// Auth wraps handlers with JWT Bearer auth and optional RBAC.
func Auth(stack *auth.Stack, strict bool, sec config.SecurityConfig, next http.Handler) http.Handler {
	return api.AuthMiddleware(stack, strict, sec.Prod, auth.PermGraphRead, next)
}
