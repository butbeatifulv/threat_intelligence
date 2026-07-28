package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
)

const (
	calderaOwner = "mitre"
	calderaRepo  = "stockpile"
)

func (u *Ingestor) ingestCalderaAbilities(ctx context.Context) error {
	base := strings.TrimSpace(os.Getenv("DS_CALDERA_BASE_PATH"))
	if base == "" {
		base = "data/abilities"
	}
	maxN := envInt("DS_MAX_CALDERA", 200)
	u.logger.Info("ingesting Caldera abilities (Stockpile)", slog.String("path", base), slog.Int("limit", maxN))

	cats, err := feeds.GitHubListDir(ctx, u.feeds, calderaOwner, calderaRepo, base)
	if err != nil {
		return fmt.Errorf("caldera list %s: %w", base, err)
	}
	n := 0
	for _, cat := range cats {
		if n >= maxN {
			break
		}
		if cat.Type != "dir" || cat.Name == ".github" {
			continue
		}
		items, err := feeds.GitHubListDir(ctx, u.feeds, calderaOwner, calderaRepo, cat.Path)
		if err != nil {
			u.logger.Debug("caldera category skip", slog.String("path", cat.Path), slog.String("err", err.Error()))
			continue
		}
		for _, it := range items {
			if n >= maxN {
				break
			}
			if it.Type != "file" || (!strings.HasSuffix(it.Name, ".yml") && !strings.HasSuffix(it.Name, ".yaml")) {
				continue
			}
			raw, unchanged, err := u.fetchBytes(ctx, calderaOwner, calderaRepo, it.Path, filepath.Join("caldera", cat.Name, strings.ReplaceAll(it.Path, "/", "__")))
			if err != nil || unchanged {
				continue
			}
			if err := u.store.UpsertCalderaRaw(ctx, it.Path, it.Name, string(raw)); err != nil {
				return err
			}
			n++
		}
	}
	u.logger.Info("caldera ingest done", slog.Int("count", n))
	return nil
}
