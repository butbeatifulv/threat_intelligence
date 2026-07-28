// Package httpmiddleware provides shared JWT Bearer + RBAC HTTP wrappers.
package httpmiddleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/butbeautifulv/veil/pkg/auth"
)

// Auth wraps handlers with JWT Bearer auth and RBAC for the given permission.
// When strict is true, only GET /health is public.
func Auth(stack *auth.Stack, strict bool, prod bool, perm string, next http.Handler) http.Handler {
	if stack == nil || !stack.Config.Enabled {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strict && r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		if strict && r.URL.Path == "/health" && r.Method == http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		raw := bearerToken(r.Header.Get("Authorization"))
		if raw == "" {
			writeAuthErr(w, http.StatusUnauthorized, auth.ErrUnauthorized, prod)
			return
		}
		sub, err := stack.Verifier.Validate(r.Context(), raw)
		if err != nil {
			writeAuthErr(w, http.StatusUnauthorized, auth.ErrUnauthorized, prod)
			return
		}
		if err := stack.Enforcer.Enforce(sub, perm); err != nil {
			writeAuthErr(w, http.StatusForbidden, err, prod)
			return
		}
		ctx := auth.WithSubject(r.Context(), sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerToken(h string) string {
	h = strings.TrimSpace(h)
	const prefix = "Bearer "
	if strings.HasPrefix(h, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(h, prefix))
	}
	return ""
}

func writeAuthErr(w http.ResponseWriter, status int, err error, prod bool) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	msg := err.Error()
	if prod {
		msg = "unauthorized"
		if status == http.StatusForbidden {
			msg = "forbidden"
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"error": msg})
}
