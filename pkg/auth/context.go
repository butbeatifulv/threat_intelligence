package auth

import "context"

type ctxKey struct{}

func WithSubject(ctx context.Context, sub *Subject) context.Context {
	return context.WithValue(ctx, ctxKey{}, sub)
}

func SubjectFromContext(ctx context.Context) (*Subject, bool) {
	v, ok := ctx.Value(ctxKey{}).(*Subject)
	return v, ok
}
