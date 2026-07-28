package usecase

import "context"

type rawPublisher interface {
	Publish(ctx context.Context, kind, contentKey string, payload any) error
}
