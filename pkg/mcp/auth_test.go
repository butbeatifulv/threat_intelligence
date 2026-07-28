package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/auth/static"
)

func testAuthStack(rbac bool, roles ...string) *auth.Stack {
	v := static.New("token", "runner", roles)
	cfg := auth.Config{
		Enabled:          true,
		RBACEnabled:      rbac,
		RoleReader:       "veil-reader",
		RoleAdmin:        "veil-admin",
		RoleEngageRunner: "veil-engage-runner",
		RoleEngageAdmin:  "veil-engage-admin",
	}
	return auth.NewStack(v, cfg)
}

func TestAuthorizeToolCall_disabled(t *testing.T) {
	stack := testAuthStack(true, "veil-engage-runner")
	stack.Config.Enabled = false
	ctx := context.Background()
	out, err := AuthorizeToolCall(ctx, stack, auth.PermEngageToolRun, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != ctx {
		t.Fatal("expected same context")
	}
}

func TestAuthorizeToolCall_nilStack(t *testing.T) {
	ctx := context.Background()
	out, err := AuthorizeToolCall(ctx, nil, auth.PermEngageToolRun, nil)
	if err != nil || out != ctx {
		t.Fatalf("got ctx=%v err=%v", out, err)
	}
}

func TestAuthorizeToolCall_contextSubject_allowed(t *testing.T) {
	stack := testAuthStack(true, "veil-engage-runner")
	sub := &auth.Subject{Sub: "u1", Roles: []string{"veil-engage-runner"}}
	ctx := auth.WithSubject(context.Background(), sub)
	out, err := AuthorizeToolCall(ctx, stack, auth.PermEngageToolRun, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := auth.SubjectFromContext(out); !ok {
		t.Fatal("subject missing from context")
	}
}

func TestAuthorizeToolCall_contextSubject_forbidden(t *testing.T) {
	stack := testAuthStack(true, "veil-reader")
	sub := &auth.Subject{Sub: "u1", Roles: []string{"veil-reader"}}
	ctx := auth.WithSubject(context.Background(), sub)
	_, err := AuthorizeToolCall(ctx, stack, auth.PermEngageToolRun, nil)
	var re *RPCError
	if !errors.As(err, &re) || re.Code != CodeAuthError || re.Message != "forbidden" {
		t.Fatalf("got %v", err)
	}
}

func TestAuthorizeToolCall_fallback_success(t *testing.T) {
	stack := testAuthStack(true, "veil-engage-runner")
	fallback := func(ctx context.Context, s *auth.Stack, _ string) (context.Context, error) {
		return auth.WithSubject(ctx, &auth.Subject{Sub: "from-fallback", Roles: []string{"veil-engage-runner"}}), nil
	}
	out, err := AuthorizeToolCall(context.Background(), stack, auth.PermEngageToolRun, fallback)
	if err != nil {
		t.Fatal(err)
	}
	sub, ok := auth.SubjectFromContext(out)
	if !ok || sub.Sub != "from-fallback" {
		t.Fatalf("subject: %+v ok=%v", sub, ok)
	}
}

func TestAuthorizeToolCall_fallback_forbidden(t *testing.T) {
	stack := testAuthStack(true, "veil-reader")
	fallback := func(ctx context.Context, s *auth.Stack, _ string) (context.Context, error) {
		return ctx, auth.ErrForbidden
	}
	_, err := AuthorizeToolCall(context.Background(), stack, auth.PermEngageToolRun, fallback)
	var re *RPCError
	if !errors.As(err, &re) || re.Message != "forbidden" {
		t.Fatalf("got %v", err)
	}
}

func TestAuthorizeToolCall_noSubject_nilFallback(t *testing.T) {
	stack := testAuthStack(true, "veil-engage-runner")
	_, err := AuthorizeToolCall(context.Background(), stack, auth.PermEngageToolRun, nil)
	var re *RPCError
	if !errors.As(err, &re) || re.Message != "unauthorized" {
		t.Fatalf("got %v", err)
	}
}
