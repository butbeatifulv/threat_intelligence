package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/ledger"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/lola/internal/repository"
	"github.com/butbeautifulv/veil/pkg/lola/domain"
)

const (
	lolbasOwner = "LOLBAS-Project"
	lolbasRepo  = "LOLBAS"
	lolbasPath  = "yml"

	gtfobinsOwner = "GTFOBins"
	gtfobinsRepo  = "GTFOBins.github.io"
	gtfobinsPath  = "_gtfobins"
)

type ScraperUsecase struct {
	repo   repository.LolaRepository
	logger *slog.Logger
	feeds  *feeds.Client
	ledger *ledger.Store
}

func NewScraperUsecase(repo repository.LolaRepository, logger *slog.Logger, fc *feeds.Client, led *ledger.Store) *ScraperUsecase {
	return &ScraperUsecase{repo: repo, logger: logger, feeds: fc, ledger: led}
}

type ghContent = feeds.GHContent

func lolaGitRef() string {
	return "master"
}

func (u *ScraperUsecase) Run(ctx context.Context) error {
	if err := u.repo.EnsureSchema(ctx); err != nil {
		return err
	}
	if err := u.IngestLOLBAS(ctx); err != nil {
		return err
	}
	if err := u.IngestGTFOBins(ctx); err != nil {
		return err
	}
	if err := u.IngestLOFTS(ctx); err != nil {
		return err
	}
	if err := u.IngestMITREEnterprise(ctx); err != nil {
		return err
	}
	return nil
}

func (u *ScraperUsecase) IngestLOLBAS(ctx context.Context) error {
	u.logger.Info("ingesting LOLBAS definitions")
	cats, err := u.githubListDir(ctx, lolbasOwner, lolbasRepo, lolbasPath)
	if err != nil {
		return err
	}
	n := 0
	for _, cat := range cats {
		if cat.Type != "dir" {
			continue
		}
		items, err := u.githubListDir(ctx, lolbasOwner, lolbasRepo, cat.Path)
		if err != nil {
			u.logger.Warn("lolbas category list failed", slog.String("path", cat.Path), slog.String("err", err.Error()))
			continue
		}
		for _, it := range items {
			if it.Type != "file" {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(it.Name), ".yml") && !strings.HasSuffix(strings.ToLower(it.Name), ".yaml") {
				continue
			}
			cacheRel := filepath.Join("lolbas", strings.ReplaceAll(it.Path, "/", "__"))
			raw, unchanged, err := u.fetchBytes(ctx, lolbasOwner, lolbasRepo, it.Path, cacheRel)
			if err != nil || unchanged {
				u.logger.Warn("lolbas file fetch failed", slog.String("path", it.Path), slog.String("err", err.Error()))
				continue
			}
			a, err := parseLOLBASYAML(raw)
			if err != nil {
				u.logger.Warn("lolbas parse failed", slog.String("path", it.Path), slog.String("err", err.Error()))
				continue
			}
			// Capture category from folder name when not present.
			if a.Category == "" {
				a.Category = strings.TrimPrefix(cat.Path, lolbasPath+"/")
			}
			if err := u.repo.UpsertArtifact(ctx, "lolbas", a); err != nil {
				return err
			}
			n++
		}
	}
	u.logger.Info("LOLBAS ingest done", slog.Int("count", n))
	return nil
}

func (u *ScraperUsecase) IngestGTFOBins(ctx context.Context) error {
	u.logger.Info("ingesting GTFOBins pages")
	items, err := u.githubListDir(ctx, gtfobinsOwner, gtfobinsRepo, gtfobinsPath)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		u.logger.Warn("GTFOBins list empty", slog.String("path", gtfobinsPath))
		return nil
	}
	n := 0
	for _, it := range items {
		if it.Type != "file" || !strings.HasSuffix(it.Name, ".md") {
			continue
		}
		cacheRel := filepath.Join("gtfobins", it.Name)
		raw, unchanged, err := u.fetchBytes(ctx, gtfobinsOwner, gtfobinsRepo, it.Path, cacheRel)
		if err != nil || unchanged {
			u.logger.Warn("gtfobins fetch failed", slog.String("name", it.Name), slog.String("err", err.Error()))
			continue
		}
		a := parseGTFOBinsMarkdown(it.Name, string(raw))
		if err := u.repo.UpsertArtifact(ctx, "gtfobins", a); err != nil {
			return err
		}
		n++
	}
	u.logger.Info("GTFOBins ingest done", slog.Int("count", n))
	return nil
}

func (u *ScraperUsecase) githubListDir(ctx context.Context, owner, repo, path string) ([]ghContent, error) {
	return feeds.GitHubListDir(ctx, u.feeds, owner, repo, path)
}

func (u *ScraperUsecase) fetchBytes(ctx context.Context, owner, repo, ghPath, cacheRel string) (body []byte, unchanged bool, err error) {
	ref := lolaGitRef()
	rawURL := feeds.GitHubRawURL(owner, repo, ref, ghPath)
	if u.ledger == nil {
		raw, err := feeds.GitHubFetchRaw(ctx, u.feeds, owner, repo, ref, ghPath)
		return raw, false, err
	}
	key := fmt.Sprintf("lola:file:%s:%s:%s", owner, repo, ghPath)
	res, err := feeds.FetchIfDue(ctx, u.feeds, u.ledger, key, "lola", rawURL, ledger.PolicyPeriodic, cacheRel, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "veil-scrape/1.0")
		return req, nil
	})
	if err != nil {
		return nil, false, err
	}
	if res.Unchanged {
		return nil, true, nil
	}
	if res.Skipped && len(res.Body) == 0 {
		return nil, false, fmt.Errorf("download %s skipped without cache", ghPath)
	}
	return res.Body, false, nil
}

func parseLOLBASYAML(raw []byte) (*domain.Artifact, error) {
	var root map[string]any
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return nil, err
	}
	a := &domain.Artifact{}
	if v, ok := root["Name"].(string); ok {
		a.Name = strings.TrimSpace(v)
	}
	if v, ok := root["Description"].(string); ok {
		a.Description = strings.TrimSpace(v)
	}
	if v, ok := root["Author"].(string); ok && v != "" {
		a.Description = strings.TrimSpace(a.Description + "\n\nAuthor: " + v)
	}
	if v, ok := root["OS"].(string); ok && v != "" {
		a.OS = []string{v}
	}
	if v, ok := root["MITRE"].(map[string]any); ok {
		if id, ok := v["ID"].(string); ok {
			a.MitreID = id
		}
	}
	if v, ok := root["Categories"].([]any); ok && len(v) > 0 {
		if s, ok := v[0].(string); ok {
			a.Category = s
		}
	}
	if v, ok := root["Full Path"].([]any); ok {
		for _, p := range v {
			if s, ok := p.(string); ok {
				a.Paths = append(a.Paths, s)
			}
		}
	}
	if v, ok := root["Commands"].([]any); ok {
		for _, c := range v {
			cm, ok := c.(map[string]any)
			if !ok {
				continue
			}
			cmd := domain.Command{}
			if s, ok := cm["Command"].(string); ok {
				cmd.Command = s
			}
			if s, ok := cm["Description"].(string); ok {
				cmd.Description = s
			}
			if s, ok := cm["Usecase"].(string); ok {
				cmd.Usecase = s
			}
			a.Commands = append(a.Commands, cmd)
		}
	}
	if det, ok := root["Detection"].(map[string]any); ok {
		if sig, ok := det["Sigma"].([]any); ok {
			for _, s := range sig {
				if str, ok := s.(string); ok {
					a.Detection.Sigma = append(a.Detection.Sigma, str)
				}
			}
		}
		if yar, ok := det["YARA"].([]any); ok {
			for _, s := range yar {
				if str, ok := s.(string); ok {
					a.Detection.Yara = append(a.Detection.Yara, str)
				}
			}
		}
	}
	if a.Name == "" {
		return nil, fmt.Errorf("missing Name")
	}
	return a, nil
}

func parseGTFOBinsMarkdown(filename, body string) *domain.Artifact {
	name := strings.TrimSuffix(filename, ".md")
	name = strings.ReplaceAll(name, "_", "/")
	a := &domain.Artifact{
		Name:        name,
		Description: firstParagraph(body),
		OS:          []string{"linux"},
		Category:    "gtfobins",
	}
	a.Commands = []domain.Command{
		{Command: "(see GTFOBins page)", Description: truncate(body, 8000)},
	}
	return a
}

func firstParagraph(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, "\n## "); idx > 0 {
		s = s[:idx]
	}
	return truncate(strings.TrimSpace(s), 4000)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "\n\n…"
}
