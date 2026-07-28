package mcpserver

import (
	"context"
)

func (s *Server) handlePlaybookProcedure(ctx context.Context, args map[string]any) (any, error) {
	id := getString(args, "id")
	if id == "" {
		return nil, errRequired("id")
	}
	spec, err := s.procedure.GetSpec(id)
	if err != nil {
		return nil, err
	}
	return toolTextResult(map[string]any{"procedure": spec})
}

func (s *Server) handlePlaybookRecommendTools(ctx context.Context, args map[string]any) (any, error) {
	id := getString(args, "id")
	if id == "" {
		tid := getString(args, "technique_id")
		if tid == "" {
			return nil, errRequired("id or technique_id")
		}
		tools := s.procedure.CatalogToolsForTechnique(tid)
		return toolTextResult(map[string]any{"technique_id": tid, "catalog_tools": tools})
	}
	tools, err := s.procedure.RecommendTools(id)
	if err != nil {
		return nil, err
	}
	return toolTextResult(map[string]any{"id": id, "catalog_tools": tools})
}

func (s *Server) handlePlaybookOntologySubdomains(ctx context.Context, args map[string]any) (any, error) {
	subs, err := s.procedure.Subdomains()
	if err != nil {
		return nil, err
	}
	return toolTextResult(map[string]any{"subdomains": subs})
}
