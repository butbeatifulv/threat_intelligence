package usecase

import (
	"context"

	"github.com/butbeautifulv/veil/knowledge/serve/internal/usecase/playbook"
	"github.com/butbeautifulv/veil/pkg/playbook/domain"
	pbindex "github.com/butbeautifulv/veil/pkg/playbook/index"
)

// ForTechnique merges index attack_ids mapping with optional Neo4j HAS_PLAYBOOK refs.
func (u *ReadUsecase) ForTechnique(ctx context.Context, techniqueID string, cat *pbindex.Catalog) (map[string]any, error) {
	fromIndex := cat.ByTechnique(techniqueID)
	graphRefs, err := u.PlaybooksForTechnique(ctx, techniqueID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"technique_id":  techniqueID,
		"index_skills":  fromIndex,
		"graph_skills":  graphRefs,
		"index_count":   len(fromIndex),
		"graph_count":   len(graphRefs),
	}, nil
}

// Summaries returns lightweight list for HTTP (no body).
func Summaries(skills []domain.SkillMeta) []map[string]any {
	out := make([]map[string]any, 0, len(skills))
	for _, s := range skills {
		out = append(out, map[string]any{
			"id":          s.ID,
			"name":        s.Name,
			"subdomain":   s.Subdomain,
			"description": s.Description,
			"attack_ids":  s.AttackIDs,
		})
	}
	return out
}

// SearchSummaries returns playbook search hits with optional score/snippet.
func SearchSummaries(hits []playbook.SearchHit) []map[string]any {
	out := make([]map[string]any, 0, len(hits))
	for _, h := range hits {
		row := map[string]any{
			"id":          h.Meta.ID,
			"name":        h.Meta.Name,
			"subdomain":   h.Meta.Subdomain,
			"description": h.Meta.Description,
			"attack_ids":  h.Meta.AttackIDs,
		}
		if h.Score > 0 {
			row["score"] = h.Score
		}
		if h.Snippet != "" {
			row["snippet"] = h.Snippet
		}
		if h.MatchType != "" {
			row["match_type"] = string(h.MatchType)
		}
		out = append(out, row)
	}
	return out
}
