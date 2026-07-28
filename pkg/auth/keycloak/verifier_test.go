package keycloak

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/butbeautifulv/veil/pkg/auth"
)

func TestNewVerifier_emptyIssuer(t *testing.T) {
	for _, iss := range []string{"", "  ", "\t"} {
		_, err := NewVerifier(context.Background(), iss, "veil-api", "veil-api")
		if err == nil {
			t.Fatalf("issuer %q: expected error", iss)
		}
	}
}

func TestNewVerifier_badJWKSURL(t *testing.T) {
	_, err := NewVerifier(context.Background(), "http://%", "veil-api", "veil-api")
	if err == nil {
		t.Fatal("expected jwks client error")
	}
}

func TestVerifier_invalidToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := NewStaticVerifier("https://kc/realms/v", "veil-api", "veil-api", &key.PublicKey)
	_, err := v.Validate(context.Background(), "not-a-jwt")
	if err != auth.ErrUnauthorized {
		t.Fatalf("got %v", err)
	}
}

func TestVerifier_expiredToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://kc/realms/v"
	v := NewStaticVerifier(issuer, "veil-api", "veil-api", &key.PublicKey)
	tok, _ := SignTestToken(key, issuer, "veil-api", "u1", nil, -time.Hour)
	_, err := v.Validate(context.Background(), tok)
	if err != auth.ErrUnauthorized {
		t.Fatalf("got %v", err)
	}
}
