package auth

import "context"

// Verifier validates a raw JWT access token and returns the subject.
type Verifier interface {
	Validate(ctx context.Context, rawJWT string) (*Subject, error)
}
