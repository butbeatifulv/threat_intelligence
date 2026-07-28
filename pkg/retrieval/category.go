package retrieval

import (
	"fmt"
	"path/filepath"
)

// CategoryRegistry holds per-category retrieval engines.
type CategoryRegistry struct {
	engines map[string]*Engine
}

// OpenCategoryRegistry loads playbook/ti/vuln indexes when present.
func OpenCategoryRegistry(cfg Config) (*CategoryRegistry, error) {
	reg := &CategoryRegistry{engines: map[string]*Engine{}}
	type catSpec struct {
		name string
		rel  string
	}
	root := cfg.RepoRoot
	specs := []catSpec{
		{"playbook", DefaultBleveRel},
		{"ti", "docs/skills-index/ti-search.bleve"},
		{"vuln", "docs/skills-index/vuln-search.bleve"},
	}
	for _, sp := range specs {
		path := filepath.Join(root, sp.rel)
		c := cfg
		c.BleveIndexPath = path
		c.SearchMode = ModeKeyword
		eng, err := NewEngine(c)
		if err != nil {
			continue
		}
		reg.engines[sp.name] = eng
	}
	if len(reg.engines) == 0 {
		return nil, fmt.Errorf("category registry: no indexes opened")
	}
	return reg, nil
}

// Search runs category-scoped search; returns skill/node ids.
func (r *CategoryRegistry) Search(category, query string, limit int) ([]SearchResult, error) {
	if r == nil {
		return nil, fmt.Errorf("category registry: nil")
	}
	eng, ok := r.engines[category]
	if !ok {
		return nil, fmt.Errorf("category registry: unknown category %q", category)
	}
	return eng.Search(SearchOpts{Query: query, Limit: limit, Mode: ModeKeyword})
}

// Close closes all engines.
func (r *CategoryRegistry) Close() error {
	if r == nil {
		return nil
	}
	var first error
	for _, e := range r.engines {
		if err := e.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}
