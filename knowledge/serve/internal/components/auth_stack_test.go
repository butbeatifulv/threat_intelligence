package components

import (
	"context"
	"testing"

	"github.com/butbeautifulv/veil/pkg/auth"
)

func TestNewAuthStack_requiresIssuer(t *testing.T) {
	_, err := newAuthStack(context.Background(), auth.Config{Enabled: true, Issuer: ""})
	if err == nil {
		t.Fatal("expected error when AUTH on without issuer")
	}
}

func TestNewAuthStack_disabled(t *testing.T) {
	st, err := newAuthStack(context.Background(), auth.Config{Enabled: false})
	if err != nil {
		t.Fatal(err)
	}
	if st.Verifier != nil {
		t.Fatal("expected nil verifier")
	}
}
