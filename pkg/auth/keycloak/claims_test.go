package keycloak

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/butbeautifulv/veil/pkg/auth"
)

func TestTokenMapClaims_notMap(t *testing.T) {
	tok := &jwt.Token{Claims: jwt.RegisteredClaims{Subject: "u"}}
	if _, ok := tokenMapClaims(tok); ok {
		t.Fatal("expected not map claims")
	}
}

func TestStaticVerifier_rejectsNonMapClaimsHook(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://kc/realms/v"
	tok, err := SignTestToken(key, issuer, "veil-api", "u1", nil, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	v := NewStaticVerifier(issuer, "veil-api", "veil-api", &key.PublicKey)
	orig := tokenMapClaims
	tokenMapClaims = func(*jwt.Token) (jwt.MapClaims, bool) { return nil, false }
	defer func() { tokenMapClaims = orig }()
	if _, err := v.Validate(context.Background(), tok); err != auth.ErrUnauthorized {
		t.Fatalf("got %v", err)
	}
}

