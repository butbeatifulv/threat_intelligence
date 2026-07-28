package keycloak

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/butbeautifulv/veil/pkg/auth"
)

func signJWKS(key *rsa.PrivateKey, kid, issuer, audience, sub, azp string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": sub, "iss": issuer, "exp": now.Add(ttl).Unix(), "iat": now.Unix(),
	}
	if audience != "" {
		claims["aud"] = audience
	}
	if azp != "" {
		claims["azp"] = azp
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = kid
	return tok.SignedString(key)
}

func rsaPublicJWKS(pub *rsa.PublicKey, kid string) []byte {
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	e := "AQAB"
	if pub.E != 65537 {
		e = base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes())
	}
	body, _ := json.Marshal(map[string]any{
		"keys": []map[string]any{{
			"kty": "RSA", "kid": kid, "alg": "RS256", "use": "sig", "n": n, "e": e,
		}},
	})
	return body
}

func TestVerifier_JWKS_wrongIssuer(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(rsaPublicJWKS(&key.PublicKey, "test-kid"))
	}))
	defer srv.Close()
	issuer := srv.URL + "/realms/veil"
	v, err := NewVerifier(context.Background(), issuer, "veil-api", "veil-api")
	if err != nil {
		t.Fatal(err)
	}
	tok, err := signJWKS(key, "test-kid", issuer+"/other", "veil-api", "u1", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.Validate(context.Background(), tok); err != auth.ErrUnauthorized {
		t.Fatalf("got %v", err)
	}
}

func TestVerifier_JWKS_emptyToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(rsaPublicJWKS(&key.PublicKey, "test-kid"))
	}))
	defer srv.Close()
	issuer := srv.URL + "/realms/veil"
	v, err := NewVerifier(context.Background(), issuer, "veil-api", "veil-api")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.Validate(context.Background(), "  "); err != auth.ErrUnauthorized {
		t.Fatalf("got %v", err)
	}
}

func TestNewVerifier_JWKS_validateOK(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	issuer := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/realms/veil/protocol/openid-connect/certs" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(rsaPublicJWKS(&key.PublicKey, "test-kid"))
	}))
	defer srv.Close()
	issuer = srv.URL + "/realms/veil"

	v, err := NewVerifier(context.Background(), issuer, "veil-api", "veil-api")
	if err != nil {
		t.Fatal(err)
	}
	tok, err := signJWKS(key, "test-kid", issuer, "veil-api", "user-1", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	sub, err := v.Validate(context.Background(), tok)
	if err != nil || sub.Sub != "user-1" {
		t.Fatalf("sub=%+v err=%v", sub, err)
	}
}

func TestVerifier_JWKS_emptySubject(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(rsaPublicJWKS(&key.PublicKey, "test-kid"))
	}))
	defer srv.Close()
	issuer := srv.URL + "/realms/veil"
	v, err := NewVerifier(context.Background(), issuer, "veil-api", "veil-api")
	if err != nil {
		t.Fatal(err)
	}
	tok, err := signJWKS(key, "test-kid", issuer, "veil-api", "", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.Validate(context.Background(), tok); err != auth.ErrUnauthorized {
		t.Fatalf("got %v", err)
	}
}

func TestVerifier_JWKS_rejectExpiredAndEmptySub(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(rsaPublicJWKS(&key.PublicKey, "test-kid"))
	}))
	defer srv.Close()
	issuer := srv.URL + "/realms/veil"
	v, err := NewVerifier(context.Background(), issuer, "veil-api", "veil-api")
	if err != nil {
		t.Fatal(err)
	}
	expTok, err := signJWKS(key, "test-kid", issuer, "veil-api", "u", "", -time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.Validate(context.Background(), expTok); err != auth.ErrUnauthorized {
		t.Fatalf("expired: %v", err)
	}
	emptySub, err := signJWKS(key, "test-kid", issuer, "veil-api", "", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.Validate(context.Background(), emptySub); err != auth.ErrUnauthorized {
		t.Fatalf("empty sub: %v", err)
	}
}

func TestValidateMapClaims_azpFallback(t *testing.T) {
	v := &Verifier{audience: "veil-api", clientID: "veil-api"}
	claims := jwt.MapClaims{
		"sub": "u1",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
		"azp": "veil-api",
	}
	sub, err := v.validateMapClaims(claims)
	if err != nil || sub.Sub != "u1" {
		t.Fatalf("azp ok: %+v err=%v", sub, err)
	}
	claims["azp"] = "other-client"
	if _, err := v.validateMapClaims(claims); err != auth.ErrUnauthorized {
		t.Fatalf("azp mismatch: %v", err)
	}
}

func TestValidateMapClaims_withAudience(t *testing.T) {
	v := &Verifier{audience: "veil-api", clientID: "c"}
	claims := jwt.MapClaims{
		"sub": "u1",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
		"aud": "veil-api",
	}
	if _, err := v.validateMapClaims(claims); err != nil {
		t.Fatal(err)
	}
}

func TestValidateMapClaims_noAudienceConfig(t *testing.T) {
	v := &Verifier{clientID: "veil-api"}
	claims := jwt.MapClaims{
		"sub": "u1",
		"exp": float64(time.Now().Add(time.Hour).Unix()),
	}
	sub, err := v.validateMapClaims(claims)
	if err != nil || sub.Sub != "u1" {
		t.Fatalf("sub=%+v err=%v", sub, err)
	}
}

func TestValidateMapClaims_expired(t *testing.T) {
	v := &Verifier{audience: "veil-api", clientID: "veil-api"}
	claims := jwt.MapClaims{
		"sub": "u1",
		"exp": float64(time.Now().Add(-time.Hour).Unix()),
	}
	if _, err := v.validateMapClaims(claims); err != auth.ErrUnauthorized {
		t.Fatalf("got %v", err)
	}
}

func TestVerifier_Validate_emptyAndWhitespace(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := NewStaticVerifier("https://kc", "veil-api", "veil-api", &key.PublicKey)
	for _, raw := range []string{"", "  "} {
		if _, err := v.Validate(context.Background(), raw); err != auth.ErrUnauthorized {
			t.Fatalf("raw=%q: %v", raw, err)
		}
	}
}

func TestVerifier_JWKS_rejectsNonMapClaimsHook(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(rsaPublicJWKS(&key.PublicKey, "kid"))
	}))
	defer srv.Close()
	issuer := srv.URL + "/realms/veil"
	v, err := NewVerifier(context.Background(), issuer, "veil-api", "veil-api")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := signJWKS(key, "kid", issuer, "veil-api", "u1", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	orig := tokenMapClaims
	tokenMapClaims = func(*jwt.Token) (jwt.MapClaims, bool) { return nil, false }
	defer func() { tokenMapClaims = orig }()
	if _, err := v.Validate(context.Background(), raw); err != auth.ErrUnauthorized {
		t.Fatalf("got %v", err)
	}
}

func TestVerifier_JWKS_wrongAudience(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(rsaPublicJWKS(&key.PublicKey, "test-kid"))
	}))
	defer srv.Close()
	issuer := srv.URL + "/realms/veil"
	v, err := NewVerifier(context.Background(), issuer, "veil-api", "veil-api")
	if err != nil {
		t.Fatal(err)
	}
	tok, err := signJWKS(key, "test-kid", issuer, "other-aud", "u1", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.Validate(context.Background(), tok); err != auth.ErrUnauthorized {
		t.Fatalf("got %v", err)
	}
}

func TestVerifier_Validate_badSigningMethod(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := NewStaticVerifier("https://kc", "aud", "client", &key.PublicKey)
	_, err := v.Validate(context.Background(), "a.b.c")
	if err != auth.ErrUnauthorized {
		t.Fatalf("got %v", err)
	}
}
