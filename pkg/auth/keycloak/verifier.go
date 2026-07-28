package keycloak

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/butbeautifulv/veil/pkg/auth"
)

// Verifier validates JWTs against Keycloak JWKS.
type Verifier struct {
	issuer   string
	audience string
	clientID string
	jwks     jwt.Keyfunc
}

// NewVerifier builds a JWKS-backed verifier for the given Keycloak realm issuer.
func NewVerifier(ctx context.Context, issuer, audience, clientID string) (*Verifier, error) {
	issuer = strings.TrimSuffix(strings.TrimSpace(issuer), "/")
	if issuer == "" {
		return nil, fmt.Errorf("keycloak issuer is empty")
	}
	jwksURL := issuer + "/protocol/openid-connect/certs"
	jwks, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("keycloak jwks: %w", err)
	}
	return &Verifier{
		issuer:   issuer,
		audience: strings.TrimSpace(audience),
		clientID: clientID,
		jwks:     jwks.Keyfunc,
	}, nil
}

func (v *Verifier) Validate(ctx context.Context, rawJWT string) (*auth.Subject, error) {
	_ = ctx
	rawJWT = strings.TrimSpace(rawJWT)
	if rawJWT == "" {
		return nil, auth.ErrUnauthorized
	}
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512"}),
		jwt.WithIssuer(v.issuer),
		jwt.WithExpirationRequired(),
	}
	if v.audience != "" {
		opts = append(opts, jwt.WithAudience(v.audience))
	}
	token, err := jwt.Parse(rawJWT, v.jwks, opts...)
	if err != nil {
		return nil, auth.ErrUnauthorized
	}
	claims, ok := tokenMapClaims(token)
	if !ok {
		return nil, auth.ErrUnauthorized
	}
	return v.validateMapClaims(claims)
}

func (v *Verifier) validateMapClaims(claims jwt.MapClaims) (*auth.Subject, error) {
	if exp, err := claims.GetExpirationTime(); err != nil || exp.Before(time.Now()) {
		return nil, auth.ErrUnauthorized
	}
	sub, _ := claims.GetSubject()
	if sub == "" {
		return nil, auth.ErrUnauthorized
	}
	// azp fallback when audience not in token
	if v.audience != "" {
		if aud, err := claims.GetAudience(); err != nil || len(aud) == 0 {
			if azp, _ := claims["azp"].(string); azp != v.audience {
				return nil, auth.ErrUnauthorized
			}
		}
	}
	return auth.SubjectFromClaims(sub, map[string]any(claims), v.clientID), nil
}
