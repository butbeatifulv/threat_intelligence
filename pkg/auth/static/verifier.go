// Package static provides a fixed-bearer verifier for local pentest/CI only (not production).
package static

import (
	"context"
	"strings"

	"github.com/butbeautifulv/veil/pkg/auth"
)

// Verifier accepts a single bearer secret and assigns pentest RBAC roles.
type Verifier struct {
	token   string
	subject string
	roles   []string
}

// New builds a verifier for AUTH_STATIC_BEARER_TOKEN (local pentest profile).
func New(token, subject string, roles []string) *Verifier {
	if subject == "" {
		subject = "pentest-runner"
	}
	if len(roles) == 0 {
		roles = []string{"veil-reader", "veil-engage-runner", "veil-engage-admin"}
	}
	return &Verifier{token: token, subject: subject, roles: roles}
}

func (v *Verifier) Validate(_ context.Context, rawJWT string) (*auth.Subject, error) {
	rawJWT = strings.TrimSpace(rawJWT)
	if rawJWT == "" || rawJWT != v.token {
		return nil, auth.ErrUnauthorized
	}
	return &auth.Subject{Sub: v.subject, Roles: append([]string(nil), v.roles...)}, nil
}
