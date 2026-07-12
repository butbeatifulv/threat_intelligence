package playbook

import (
	"strings"

	"github.com/butbeautifulv/veil/pkg/playbook/domain"
	pbindex "github.com/butbeautifulv/veil/pkg/playbook/index"
	"github.com/butbeautifulv/veil/pkg/retrieval"
)

// SearchHit is enriched playbook search result.
type SearchHit struct {
	Meta      domain.SkillMeta
	Score     float64
	Snippet   string
	MatchType retrieval.MatchType
}

// Service exposes read-only cybersecurity playbook skills from the generated index.
type Service struct {
	cat      *pbindex.Catalog
	engine   *retrieval.Engine
	cfg      retrieval.Config
	useBleve bool
}

// NewService loads the default catalog and optional retrieval engine.
func NewService() (*Service, error) {
	cat, err := pbindex.Default()
	if err != nil {
		return nil, err
	}
	root, err := pbindex.RepoRoot()
	if err != nil {
		root, err = retrieval.VeilRoot()
	}
	if err != nil {
		return &Service{cat: cat}, nil
	}
	cfg := retrieval.ConfigFromEnv(root)
	svc := &Service{cat: cat, cfg: cfg}
	if strings.ToLower(cfg.SearchEngine) == "legacy" {
		return svc, nil
	}
	eng, err := retrieval.NewEngine(cfg)
	if err != nil {
		return svc, nil
	}
	retrieval.WireEmbedder(eng, cfg)
	if cfg.RerankEnabled && cfg.RerankURL != "" {
		eng.SetReranker(retrieval.NewRerankClient(cfg.RerankURL))
	}
	svc.engine = eng
	svc.useBleve = true
	return svc, nil
}

func (s *Service) IndexMeta() domain.IndexFile {
	return s.cat.Meta()
}

func (s *Service) Search(query, subdomain string, limit int) []domain.SkillMeta {
	hits := s.SearchEnriched(query, subdomain, limit, "")
	out := make([]domain.SkillMeta, 0, len(hits))
	for _, h := range hits {
		out = append(out, h.Meta)
	}
	return out
}

// SearchEnriched returns skill metadata with optional score and snippet.
func (s *Service) SearchEnriched(query, subdomain string, limit int, mode string) []SearchHit {
	if s.useBleve && s.engine != nil {
		opts := retrieval.SearchOpts{
			Query:     query,
			Subdomain: subdomain,
			Limit:     limit,
		}
		if mode != "" {
			opts.Mode = retrieval.SearchMode(strings.ToLower(mode))
		}
		res, err := s.engine.Search(opts)
		if err == nil {
			return s.hitsFromResults(res)
		}
	}
	legacy := s.cat.Search(query, subdomain, limit)
	out := make([]SearchHit, 0, len(legacy))
	for _, m := range legacy {
		out = append(out, SearchHit{
			Meta:      m,
			MatchType: retrieval.MatchKeyword,
		})
	}
	return out
}

func (s *Service) hitsFromResults(res []retrieval.SearchResult) []SearchHit {
	out := make([]SearchHit, 0, len(res))
	for _, r := range res {
		meta, ok := s.cat.Get(r.SkillID)
		if !ok {
			continue
		}
		out = append(out, SearchHit{
			Meta:      meta,
			Score:     r.Score,
			Snippet:   r.Snippet,
			MatchType: r.MatchType,
		})
	}
	return out
}

// SearchMode returns configured default search mode.
func (s *Service) SearchMode() string {
	return string(s.cfg.SearchMode)
}

func (s *Service) Get(id string) (domain.SkillDetail, error) {
	return s.cat.ReadBody(id)
}

func (s *Service) ByTechnique(techniqueID string) []domain.SkillMeta {
	return s.cat.ByTechnique(techniqueID)
}

// Catalog returns the underlying index (for merge with graph query).
func (s *Service) Catalog() *pbindex.Catalog {
	return s.cat
}

// Close releases retrieval resources.
func (s *Service) Close() error {
	if s.engine != nil {
		return s.engine.Close()
	}
	return nil
}
