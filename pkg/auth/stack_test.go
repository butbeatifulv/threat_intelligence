package auth

import (
	"context"
	"testing"
)

type stubVerifier struct{}

func (stubVerifier) Validate(context.Context, string) (*Subject, error) {
	return nil, ErrUnauthorized
}

func TestNewStack(t *testing.T) {
	cfg := Config{
		Enabled:          true,
		RBACEnabled:      true,
		RoleReader:       "veil-reader",
		RoleAdmin:        "veil-admin",
		RoleEngageRunner: "veil-engage-runner",
		RoleEngageAdmin:  "veil-engage-admin",
	}
	v := stubVerifier{}
	stack := NewStack(v, cfg)

	if stack.Config != cfg {
		t.Fatalf("Config: %+v", stack.Config)
	}
	if stack.Verifier != v {
		t.Fatal("Verifier not set")
	}
	if stack.Enforcer == nil {
		t.Fatal("Enforcer is nil")
	}
	if err := stack.Enforcer.Enforce(&Subject{Roles: []string{"veil-reader"}}, PermGraphRead); err != nil {
		t.Fatalf("enforcer wired: %v", err)
	}
}
