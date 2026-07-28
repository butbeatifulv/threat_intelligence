package auth

import (
	"context"
	"testing"
)

func TestWithSubject_SubjectFromContext(t *testing.T) {
	sub := &Subject{Sub: "user-1", Email: "u@example.com", Roles: []string{"veil-reader"}}
	ctx := WithSubject(context.Background(), sub)

	got, ok := SubjectFromContext(ctx)
	if !ok {
		t.Fatal("expected subject in context")
	}
	if got != sub {
		t.Fatalf("subject pointer mismatch: got %+v", got)
	}
}

func TestSubjectFromContext_missing(t *testing.T) {
	_, ok := SubjectFromContext(context.Background())
	if ok {
		t.Fatal("expected no subject")
	}
}

func TestSubjectFromContext_wrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxKey{}, "not-a-subject")
	_, ok := SubjectFromContext(ctx)
	if ok {
		t.Fatal("expected false for wrong value type")
	}
}
