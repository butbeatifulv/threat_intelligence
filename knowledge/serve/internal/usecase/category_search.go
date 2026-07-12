package usecase

import (
	"github.com/butbeautifulv/veil/pkg/retrieval"
)

// ReadUsecase hybrid search extension.
func (u *ReadUsecase) initCategoryRegistry() {
	if u == nil || u.categoryReg != nil {
		return
	}
	root, err := retrieval.VeilRoot()
	if err != nil {
		return
	}
	cfg := retrieval.ConfigFromEnv(root)
	reg, err := retrieval.OpenCategoryRegistry(cfg)
	if err != nil {
		return
	}
	u.categoryReg = reg
}

// CategoryHybridSearch uses Bleve indexes for ti/vuln when available.
func (u *ReadUsecase) CategoryHybridSearch(category, query string, limit int) ([]map[string]any, bool) {
	if category != "ti" && category != "vuln" {
		return nil, false
	}
	u.initCategoryRegistry()
	if u.categoryReg == nil {
		return nil, false
	}
	res, err := u.categoryReg.Search(category, query, limit)
	if err != nil {
		return nil, false
	}
	out := make([]map[string]any, 0, len(res))
	for _, r := range res {
		out = append(out, map[string]any{
			"id":      r.SkillID,
			"score":   r.Score,
			"snippet": r.Snippet,
		})
	}
	return out, true
}
