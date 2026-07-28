package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
)

func (u *Ingestor) UpsertActor(ctx context.Context, a domain.Actor) error {
	if strings.TrimSpace(a.Name) == "" {
		return fmt.Errorf("actor requires name")
	}
	return u.repo.UpsertActor(ctx, a)
}

func (u *Ingestor) UpsertReport(ctx context.Context, r domain.Report) error {
	if strings.TrimSpace(r.Title) == "" || strings.TrimSpace(r.Link) == "" {
		return fmt.Errorf("report requires title and link")
	}
	return u.repo.UpsertReport(ctx, r)
}
