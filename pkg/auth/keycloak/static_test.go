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

func TestStaticVerifier_roles(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	issuer := "https://keycloak.example/realms/veil"
	aud := "veil-api"
	tok, err := SignTestToken(key, issuer, aud, "user-1", []string{"veil-reader"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	v := NewStaticVerifier(issuer, aud, "veil-api", &key.PublicKey)
	sub, err := v.Validate(context.Background(), tok)
	if err != nil {
		t.Fatal(err)
	}
	if sub.Sub != "user-1" {
		t.Fatalf("sub: %s", sub.Sub)
	}
	if !sub.HasRole("veil-reader") {
		t.Fatalf("roles: %v", sub.Roles)
	}
	e := auth.NewEnforcer(auth.Config{
		Enabled:     true,
		RBACEnabled: true,
		RoleReader:  "veil-reader",
	})
	if err := e.Enforce(sub, auth.PermGraphRead); err != nil {
		t.Fatal(err)
	}
}

func TestStaticVerifier_wrongSigningMethod(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := NewStaticVerifier("https://kc", "veil-api", "veil-api", &key.PublicKey)
	// HS256 token with arbitrary secret
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "u", "iss": "https://kc", "exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.Validate(context.Background(), raw); err != auth.ErrUnauthorized {
		t.Fatalf("got %v", err)
	}
}

func TestStaticVerifier_noAudience(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://kc/realms/v"
	tok, err := SignTestToken(key, issuer, "", "u1", nil, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	v := NewStaticVerifier(issuer, "", "veil-api", &key.PublicKey)
	sub, err := v.Validate(context.Background(), tok)
	if err != nil || sub.Sub != "u1" {
		t.Fatalf("sub=%+v err=%v", sub, err)
	}
}

func TestStaticVerifier_emptySubject(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://kc/realms/v"
	tok, err := SignTestToken(key, issuer, "veil-api", "", nil, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	v := NewStaticVerifier(issuer, "veil-api", "veil-api", &key.PublicKey)
	if _, err := v.Validate(context.Background(), tok); err != auth.ErrUnauthorized {
		t.Fatalf("got %v", err)
	}
}

func TestStaticVerifier_invalid(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := NewStaticVerifier("https://kc/realms/v", "veil-api", "veil-api", &key.PublicKey)
	if _, err := v.Validate(context.Background(), ""); err != auth.ErrUnauthorized {
		t.Fatalf("got %v", err)
	}
}
