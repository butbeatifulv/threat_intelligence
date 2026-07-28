package keycloak

import (
	"context"
	"crypto/rsa"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/butbeautifulv/veil/pkg/auth"
)

// StaticVerifier validates JWTs signed with a fixed RSA public key (tests).
type StaticVerifier struct {
	issuer   string
	audience string
	clientID string
	key      *rsa.PublicKey
}

func NewStaticVerifier(issuer, audience, clientID string, pub *rsa.PublicKey) *StaticVerifier {
	return &StaticVerifier{
		issuer:   strings.TrimSuffix(strings.TrimSpace(issuer), "/"),
		audience: audience,
		clientID: clientID,
		key:      pub,
	}
}

func (v *StaticVerifier) Validate(ctx context.Context, rawJWT string) (*auth.Subject, error) {
	_ = ctx
	rawJWT = strings.TrimSpace(rawJWT)
	if rawJWT == "" {
		return nil, auth.ErrUnauthorized
	}
	opts := []jwt.ParserOption{jwt.WithIssuer(v.issuer), jwt.WithExpirationRequired()}
	if v.audience != "" {
		opts = append(opts, jwt.WithAudience(v.audience))
	}
	token, err := jwt.Parse(rawJWT, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return v.key, nil
	}, opts...)
	if err != nil {
		return nil, auth.ErrUnauthorized
	}
	claims, ok := tokenMapClaims(token)
	if !ok {
		return nil, auth.ErrUnauthorized
	}
	sub, _ := claims.GetSubject()
	if sub == "" {
		return nil, auth.ErrUnauthorized
	}
	return auth.SubjectFromClaims(sub, map[string]any(claims), v.clientID), nil
}

// SignTestToken issues a JWT for tests (RS256).
func SignTestToken(priv *rsa.PrivateKey, issuer, audience, sub string, roles []string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": sub,
		"iss": issuer,
		"exp": now.Add(ttl).Unix(),
		"iat": now.Unix(),
		"realm_access": map[string]any{
			"roles": roles,
		},
	}
	if audience != "" {
		claims["aud"] = audience
	}
	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return t.SignedString(priv)
}
