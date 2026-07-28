package static

import (
	"context"
	"testing"

	"github.com/butbeautifulv/veil/pkg/auth"
)

func TestNew_defaults(t *testing.T) {
	v := New("tok", "", nil)
	sub, err := v.Validate(context.Background(), "tok")
	if err != nil {
		t.Fatal(err)
	}
	if sub.Sub != "pentest-runner" {
		t.Fatalf("sub %q", sub.Sub)
	}
	if len(sub.Roles) < 3 {
		t.Fatalf("roles %v", sub.Roles)
	}
}

func TestVerifier_Validate(t *testing.T) {
	v := New("secret-token", "u", nil)
	_, err := v.Validate(context.Background(), "wrong")
	if err != auth.ErrUnauthorized {
		t.Fatalf("want unauthorized, got %v", err)
	}
	sub, err := v.Validate(context.Background(), "secret-token")
	if err != nil || sub.Sub != "u" {
		t.Fatalf("sub=%+v err=%v", sub, err)
	}
}
