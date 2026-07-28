package mcpserver

import (
	"context"

	"github.com/butbeautifulv/veil/knowledge/serve/internal/usecase"
)

func (s *Server) handlePlaybookSearch(ctx context.Context, args map[string]any) (any, error) {
	q := getString(args, "query")
	sub := getString(args, "subdomain")
	limit := getInt(args, "limit", 15)
	mode := getString(args, "mode")
	hits := s.playbook.SearchEnriched(q, sub, limit, mode)
	return toolTextResult(map[string]any{
		"query":       q,
		"subdomain":   sub,
		"search_mode": s.playbook.SearchMode(),
		"skills":      usecase.SearchSummaries(hits),
		"count":       len(hits),
	})
}

func (s *Server) handlePlaybookGet(ctx context.Context, args map[string]any) (any, error) {
	id := getString(args, "id")
	if id == "" {
		return nil, errRequired("id")
	}
	detail, err := s.playbook.Get(id)
	if err != nil {
		return nil, err
	}
	return toolTextResult(map[string]any{
		"skill": detail,
	})
}

func (s *Server) handlePlaybookForTechnique(ctx context.Context, args map[string]any) (any, error) {
	tid := getString(args, "technique_id")
	if tid == "" {
		return nil, errRequired("technique_id")
	}
	out, err := s.uc.ForTechnique(ctx, tid, s.playbook.Catalog())
	if err != nil {
		return nil, err
	}
	return toolTextResult(out)
}

func errRequired(field string) error {
	return &requiredError{field: field}
}

type requiredError struct{ field string }

func (e *requiredError) Error() string {
	return e.field + " is required"
}
