package playbookseed

import (
	"context"
	"log/slog"
)

// CatalogMeta is the skills index metadata needed for seeding.
type CatalogMeta struct {
	Skills []SkillSeed
}

// Run seeds CyberSkill nodes and HAS_PLAYBOOK edges from catalog metadata.
func Run(ctx context.Context, repo Repository, meta CatalogMeta, log *slog.Logger) (int, int, error) {
	if log == nil {
		log = slog.Default()
	}
	if err := repo.EnsureConstraints(ctx); err != nil {
		return 0, 0, err
	}
	linked := 0
	for _, s := range meta.Skills {
		if err := repo.UpsertSkill(ctx, s); err != nil {
			log.Warn("skill upsert failed", slog.String("id", s.ID), slog.Any("err", err))
			continue
		}
		for _, tid := range s.AttackIDs {
			if err := repo.LinkAttackTechnique(ctx, tid, s.ID); err != nil {
				continue
			}
			linked++
		}
	}
	return len(meta.Skills), linked, nil
}
