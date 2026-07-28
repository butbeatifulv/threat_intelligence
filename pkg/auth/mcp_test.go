package auth_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/auth/keycloak"
)

func TestAuthorizeMCP_contextSubject(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://kc/realms/v"
	v := keycloak.NewStaticVerifier(issuer, "veil-api", "veil-api", &key.PublicKey)
	stack := auth.NewStack(v, auth.Config{Enabled: true, RBACEnabled: true, RoleReader: "veil-reader"})
	sub := &auth.Subject{Sub: "u1", Roles: []string{"veil-reader"}}
	ctx := auth.WithSubject(context.Background(), sub)

	out, err := auth.AuthorizeMCP(ctx, stack, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := auth.SubjectFromContext(out); !ok {
		t.Fatal("expected subject in context")
	}
}

func TestAuthorizeMCP_explicitToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://kc/realms/v"
	v := keycloak.NewStaticVerifier(issuer, "veil-api", "veil-api", &key.PublicKey)
	tok, _ := keycloak.SignTestToken(key, issuer, "veil-api", "u1", []string{"veil-reader"}, time.Hour)
	stack := auth.NewStack(v, auth.Config{Enabled: true, RBACEnabled: true, RoleReader: "veil-reader"})
	out, err := auth.AuthorizeMCP(context.Background(), stack, tok)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := auth.SubjectFromContext(out); !ok {
		t.Fatal("expected subject")
	}
}

func TestAuthorizeEngageMCP_explicitToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://kc/realms/v"
	v := keycloak.NewStaticVerifier(issuer, "veil-api", "veil-api", &key.PublicKey)
	tok, _ := keycloak.SignTestToken(key, issuer, "veil-api", "u1", []string{"veil-engage-runner"}, time.Hour)
	stack := auth.NewStack(v, auth.Config{
		Enabled:          true,
		RBACEnabled:      true,
		RoleEngageRunner: "veil-engage-runner",
	})
	out, err := auth.AuthorizeEngageMCP(context.Background(), stack, tok)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := auth.SubjectFromContext(out); !ok {
		t.Fatal("expected subject")
	}
}

func TestAuthorizeMCP_envToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://kc/realms/v"
	v := keycloak.NewStaticVerifier(issuer, "veil-api", "veil-api", &key.PublicKey)
	tok, _ := keycloak.SignTestToken(key, issuer, "veil-api", "u1", []string{"veil-reader"}, time.Hour)
	stack := auth.NewStack(v, auth.Config{
		Enabled:        true,
		RBACEnabled:    true,
		RoleReader:     "veil-reader",
		MCPAccessToken: tok,
	})
	_, err := auth.AuthorizeMCP(context.Background(), stack, "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestAuthorizeMCP_disabled_nilStack(t *testing.T) {
	ctx := context.Background()
	out, err := auth.AuthorizeMCP(ctx, nil, "")
	if err != nil || out != ctx {
		t.Fatalf("nil stack: err=%v", err)
	}
	stack := auth.NewStack(nil, auth.Config{Enabled: false})
	out, err = auth.AuthorizeMCP(ctx, stack, "")
	if err != nil || out != ctx {
		t.Fatalf("disabled: err=%v", err)
	}
}

func TestAuthorizeMCP_emptyToken(t *testing.T) {
	stack := auth.NewStack(nil, auth.Config{Enabled: true})
	_, err := auth.AuthorizeMCP(context.Background(), stack, "")
	if err != auth.ErrUnauthorized {
		t.Fatalf("nil verifier: %v", err)
	}
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://kc/realms/v"
	v := keycloak.NewStaticVerifier(issuer, "veil-api", "veil-api", &key.PublicKey)
	stack = auth.NewStack(v, auth.Config{Enabled: true})
	_, err = auth.AuthorizeMCP(context.Background(), stack, "")
	if err != auth.ErrUnauthorized {
		t.Fatalf("empty configured token: %v", err)
	}
}

func TestAuthorizeEngageMCP_contextSubject(t *testing.T) {
	stack := auth.NewStack(nil, auth.Config{
		Enabled:          true,
		RBACEnabled:      true,
		RoleEngageRunner: "veil-engage-runner",
	})
	sub := &auth.Subject{Sub: "u1", Roles: []string{"veil-engage-runner"}}
	ctx := auth.WithSubject(context.Background(), sub)
	out, err := auth.AuthorizeEngageMCP(ctx, stack, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := auth.SubjectFromContext(out); !ok {
		t.Fatal("expected subject in context")
	}
}

func TestAuthorizeEngageMCP_envToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://kc/realms/v"
	v := keycloak.NewStaticVerifier(issuer, "veil-api", "veil-api", &key.PublicKey)
	tok, _ := keycloak.SignTestToken(key, issuer, "veil-api", "u1", []string{"veil-engage-runner"}, time.Hour)
	stack := auth.NewStack(v, auth.Config{
		Enabled:          true,
		RBACEnabled:      true,
		RoleEngageRunner: "veil-engage-runner",
		MCPAccessToken:   tok,
	})
	out, err := auth.AuthorizeEngageMCP(context.Background(), stack, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := auth.SubjectFromContext(out); !ok {
		t.Fatal("expected subject from token path")
	}
}

func TestAuthorizeEngageMCP_forbidden(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://kc/realms/v"
	v := keycloak.NewStaticVerifier(issuer, "veil-api", "veil-api", &key.PublicKey)
	tok, _ := keycloak.SignTestToken(key, issuer, "veil-api", "u1", []string{"veil-reader"}, time.Hour)
	stack := auth.NewStack(v, auth.Config{
		Enabled:        true,
		RBACEnabled:    true,
		RoleEngageRunner: "veil-engage-runner",
		MCPAccessToken: tok,
	})
	_, err := auth.AuthorizeEngageMCP(context.Background(), stack, "")
	if err != auth.ErrForbidden {
		t.Fatalf("got %v", err)
	}
}

func TestAuthorizeEngageMCP_disabledAndEmptyToken(t *testing.T) {
	ctx := context.Background()
	out, err := auth.AuthorizeEngageMCP(ctx, nil, "")
	if err != nil || out != ctx {
		t.Fatalf("nil stack: err=%v", err)
	}
	stack := auth.NewStack(nil, auth.Config{Enabled: true})
	if _, err := auth.AuthorizeEngageMCP(ctx, stack, ""); err != auth.ErrUnauthorized {
		t.Fatalf("nil verifier: %v", err)
	}
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://kc/realms/v"
	v := keycloak.NewStaticVerifier(issuer, "veil-api", "veil-api", &key.PublicKey)
	stack = auth.NewStack(v, auth.Config{Enabled: true})
	if _, err := auth.AuthorizeEngageMCP(ctx, stack, ""); err != auth.ErrUnauthorized {
		t.Fatalf("empty configured token: %v", err)
	}
}

func TestAuthorizeEngageMCP_contextForbidden(t *testing.T) {
	stack := auth.NewStack(nil, auth.Config{
		Enabled:          true,
		RBACEnabled:      true,
		RoleEngageRunner: "veil-engage-runner",
	})
	sub := &auth.Subject{Sub: "u1", Roles: []string{"veil-reader"}}
	ctx := auth.WithSubject(context.Background(), sub)
	_, err := auth.AuthorizeEngageMCP(ctx, stack, "")
	if err != auth.ErrForbidden {
		t.Fatalf("got %v", err)
	}
}

func TestAuthorizeMCP_invalidToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://kc/realms/v"
	v := keycloak.NewStaticVerifier(issuer, "veil-api", "veil-api", &key.PublicKey)
	stack := auth.NewStack(v, auth.Config{Enabled: true, MCPAccessToken: "not-a-jwt"})
	_, err := auth.AuthorizeMCP(context.Background(), stack, "")
	if err != auth.ErrUnauthorized {
		t.Fatalf("got %v", err)
	}
}

func TestAuthorizeMCP_contextForbidden(t *testing.T) {
	stack := auth.NewStack(nil, auth.Config{Enabled: true, RBACEnabled: true, RoleReader: "veil-reader"})
	sub := &auth.Subject{Sub: "u1", Roles: []string{"other"}}
	ctx := auth.WithSubject(context.Background(), sub)
	_, err := auth.AuthorizeMCP(ctx, stack, "")
	if err != auth.ErrForbidden {
		t.Fatalf("got %v", err)
	}
}

func TestAuthorizeEngageMCP_invalidToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://kc/realms/v"
	v := keycloak.NewStaticVerifier(issuer, "veil-api", "veil-api", &key.PublicKey)
	stack := auth.NewStack(v, auth.Config{Enabled: true, MCPAccessToken: "not-a-jwt"})
	if _, err := auth.AuthorizeEngageMCP(context.Background(), stack, ""); err != auth.ErrUnauthorized {
		t.Fatalf("env token: %v", err)
	}
	if _, err := auth.AuthorizeEngageMCP(context.Background(), stack, "not-a-jwt"); err != auth.ErrUnauthorized {
		t.Fatalf("explicit token: %v", err)
	}
}

func TestAuthorizeMCP_forbidden(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://kc/realms/v"
	v := keycloak.NewStaticVerifier(issuer, "veil-api", "veil-api", &key.PublicKey)
	tok, _ := keycloak.SignTestToken(key, issuer, "veil-api", "u1", []string{"other"}, time.Hour)
	stack := auth.NewStack(v, auth.Config{
		Enabled:        true,
		RBACEnabled:    true,
		RoleReader:     "veil-reader",
		MCPAccessToken: tok,
	})
	_, err := auth.AuthorizeMCP(context.Background(), stack, "")
	if err != auth.ErrForbidden {
		t.Fatalf("got %v", err)
	}
}
