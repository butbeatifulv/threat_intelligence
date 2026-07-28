package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/ti/domain"
	"github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/ti/jsonl"
	"github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/ti/repository"
)

type Ingestor struct {
	repo   repository.GraphRepository
	logger *slog.Logger
}

func NewIngestor(repo repository.GraphRepository, logger *slog.Logger) *Ingestor {
	return &Ingestor{repo: repo, logger: logger}
}

func (u *Ingestor) IngestJSONL(ctx context.Context, stream *jsonl.Stream) error {
	if err := u.repo.EnsureSchema(ctx); err != nil {
		return err
	}

	processed := 0
	skipped := 0

	_, err := stream.Walk(ctx, func(env jsonl.Envelope) error {
		switch {
		case env.IOC != nil:
			ok, err := u.UpsertIOCFile(ctx, *env.IOC)
			if err != nil {
				return err
			}
			if ok {
				processed++
			} else {
				skipped++
			}
			return nil

		case env.Campaign != nil:
			if err := u.UpsertCampaign(ctx, *env.Campaign); err != nil {
				return err
			}
			processed++
			return nil

		case env.Cluster != nil:
			if err := u.UpsertCluster(ctx, *env.Cluster); err != nil {
				return err
			}
			processed++
			return nil

		case env.Actor != nil:
			if err := u.UpsertActor(ctx, *env.Actor); err != nil {
				return err
			}
			processed++
			return nil

		case env.Report != nil:
			if err := u.UpsertReport(ctx, *env.Report); err != nil {
				return err
			}
			processed++
			return nil
		default:
			skipped++
			return nil
		}
	})
	if err != nil {
		return err
	}

	u.logger.Info("ingest finished",
		slog.Int("processed", processed),
		slog.Int("skipped", skipped),
	)
	return nil
}

// IngestOne applies a single JSONL-shaped envelope (offline / TIJSONLRecord path).
func (u *Ingestor) IngestOne(ctx context.Context, env jsonl.Envelope) error {
	switch {
	case env.IOC != nil:
		_, err := u.UpsertIOCFile(ctx, *env.IOC)
		return err
	case env.Campaign != nil:
		return u.UpsertCampaign(ctx, *env.Campaign)
	case env.Cluster != nil:
		return u.UpsertCluster(ctx, *env.Cluster)
	case env.Actor != nil:
		return u.UpsertActor(ctx, *env.Actor)
	case env.Report != nil:
		return u.UpsertReport(ctx, *env.Report)
	default:
		return nil
	}
}

// UpsertIOC applies a NED-normalized IOC; Neo4j node id comes from commit idempotency key.
func (u *Ingestor) UpsertIOC(ctx context.Context, idempotencyKey string, i domain.IOC) (bool, error) {
	if strings.TrimSpace(i.Value) == "" {
		return false, nil
	}
	id, err := commit.IOCNodeID(idempotencyKey)
	if err != nil {
		return false, err
	}
	return true, u.repo.UpsertIOC(ctx, id, i)
}

// UpsertIOCFile is for offline JSONL ingest (no commit envelope); uses IOC.NodeID hash.
func (u *Ingestor) UpsertIOCFile(ctx context.Context, i domain.IOC) (bool, error) {
	if strings.TrimSpace(i.Value) == "" {
		return false, nil
	}
	return true, u.repo.UpsertIOC(ctx, i.NodeID(), i)
}

func (u *Ingestor) UpsertCampaign(ctx context.Context, c domain.Campaign) error {
	if c.ID == "" || c.Name == "" {
		return fmt.Errorf("campaign requires id and name")
	}
	if err := u.repo.UpsertCampaign(ctx, c); err != nil {
		return err
	}
	for _, actor := range c.Actors {
		if actor == "" {
			continue
		}
		if err := u.repo.LinkCampaignActor(ctx, c.ID, actor, actor); err != nil {
			return err
		}
	}
	for _, i := range c.IOCs {
		if strings.TrimSpace(i.Value) == "" {
			continue
		}
		id := i.NodeID()
		if err := u.repo.UpsertIOC(ctx, id, i); err != nil {
			return err
		}
		if err := u.repo.LinkCampaignIOC(ctx, c.ID, id); err != nil {
			return err
		}
	}
	return nil
}

func (u *Ingestor) UpsertCluster(ctx context.Context, cl domain.Cluster) error {
	if cl.ID == "" || cl.Name == "" {
		return fmt.Errorf("cluster requires id and name")
	}
	if err := u.repo.UpsertCluster(ctx, cl); err != nil {
		return err
	}
	for _, c := range cl.Campaigns {
		if err := u.UpsertCampaign(ctx, c); err != nil {
			return err
		}
		if err := u.repo.LinkClusterCampaign(ctx, cl.ID, c.ID); err != nil {
			return err
		}
	}
	return nil
}
