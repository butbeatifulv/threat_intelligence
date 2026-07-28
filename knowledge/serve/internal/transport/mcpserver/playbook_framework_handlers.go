package mcpserver

import (
	"context"
	"encoding/json"
)

func (s *Server) handlePlaybookFramework(ctx context.Context, args map[string]any) (any, error) {
	which := getString(args, "framework")
	if which == "" {
		which = "mitre"
	}
	switch which {
	case "mitre", "mitre-layer":
		raw, err := s.framework.RawMitreLayerJSON()
		if err != nil {
			return nil, err
		}
		var doc any
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, err
		}
		return toolTextResult(map[string]any{"framework": "mitre", "layer": doc})
	case "coverage":
		out, err := s.framework.MitreCoverage()
		if err != nil {
			return nil, err
		}
		return toolTextResult(out)
	case "docs":
		docs, err := s.framework.ListMappingDocs()
		if err != nil {
			return nil, err
		}
		return toolTextResult(map[string]any{"files": docs})
	default:
		return toolTextResult(map[string]any{
			"error":   "unknown framework",
			"allowed": []string{"mitre", "coverage", "docs"},
		})
	}
}

func (s *Server) handlePlaybookSubdomains(ctx context.Context, args map[string]any) (any, error) {
	meta := s.playbook.IndexMeta()
	return toolTextResult(map[string]any{
		"subdomain_counts": meta.SubdomainCounts,
		"skill_count":      meta.SkillCount,
	})
}
