package retrieval

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pbindex "github.com/butbeautifulv/veil/pkg/playbook/index"
)

const DefaultEvalRel = "docs/skills-index/search-eval.json"

// EvalQuery is one golden search case.
type EvalQuery struct {
	ID             string   `json:"id"`
	Query          string   `json:"query"`
	ExpectedIDs    []string `json:"expected_ids"`
	Type           string   `json:"type"`
	Subdomain      string   `json:"subdomain,omitempty"`
	MatchTechnique string   `json:"match_technique,omitempty"`
	ExpectEmpty    bool     `json:"expect_empty,omitempty"`
}

// EvalFile is the search-eval.json document.
type EvalFile struct {
	SchemaVersion int         `json:"schema_version"`
	Queries       []EvalQuery `json:"queries"`
}

// EvalMetrics holds aggregate search quality metrics.
type EvalMetrics struct {
	Total       int
	Scored      int
	RecallAt5   float64
	MRR         float64
	PerQuery    []EvalQueryResult
}

// EvalQueryResult is per-query eval output.
type EvalQueryResult struct {
	ID        string  `json:"id"`
	Query     string  `json:"query"`
	HitAt5    bool    `json:"hit_at_5"`
	RR        float64 `json:"reciprocal_rank"`
	TopID     string  `json:"top_id,omitempty"`
	Skipped   bool    `json:"skipped,omitempty"`
	SkipReason string `json:"skip_reason,omitempty"`
}

// SearchFunc runs playbook search (legacy or retrieval engine).
type SearchFunc func(query, subdomain string, limit int) []string

// LoadEvalFile reads golden queries from repo-relative path.
func LoadEvalFile(repoRoot, rel string) (EvalFile, error) {
	if rel == "" {
		rel = DefaultEvalRel
	}
	path := rel
	if !filepath.IsAbs(path) {
		path = filepath.Join(repoRoot, rel)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return EvalFile{}, fmt.Errorf("eval: read %s: %w", path, err)
	}
	var f EvalFile
	if err := json.Unmarshal(raw, &f); err != nil {
		return EvalFile{}, fmt.Errorf("eval: decode: %w", err)
	}
	return f, nil
}

// EvalLegacy runs metrics against Catalog.Search (baseline).
func EvalLegacy(cat *pbindex.Catalog, file EvalFile, limit int) EvalMetrics {
	return Eval(file, func(query, subdomain string, lim int) []string {
		hits := cat.Search(query, subdomain, lim)
		ids := make([]string, 0, len(hits))
		for _, h := range hits {
			ids = append(ids, h.ID)
		}
		return ids
	}, limit)
}

// Eval runs recall@5 and MRR for a search function.
func Eval(file EvalFile, search SearchFunc, limit int) EvalMetrics {
	if limit <= 0 {
		limit = 5
	}
	m := EvalMetrics{Total: len(file.Queries), PerQuery: make([]EvalQueryResult, 0, len(file.Queries))}
	var recallSum, mrrSum float64
	var scored int
	for _, q := range file.Queries {
		res := evalOne(q, search, limit)
		m.PerQuery = append(m.PerQuery, res)
		if res.Skipped {
			continue
		}
		scored++
		if res.HitAt5 {
			recallSum++
		}
		mrrSum += res.RR
	}
	m.Scored = scored
	if scored > 0 {
		m.RecallAt5 = recallSum / float64(scored)
		m.MRR = mrrSum / float64(scored)
	}
	return m
}

func evalOne(q EvalQuery, search SearchFunc, limit int) EvalQueryResult {
	res := EvalQueryResult{ID: q.ID, Query: q.Query}
	if q.MatchTechnique != "" {
		res.Skipped = true
		res.SkipReason = "technique lookup not via search"
		return res
	}
	if len(q.ExpectedIDs) == 0 && !q.ExpectEmpty {
		res.Skipped = true
		res.SkipReason = "no expected_ids (open-ended query)"
		return res
	}
	sub := q.Subdomain
	ids := search(q.Query, sub, limit)
	if len(ids) > 0 {
		res.TopID = ids[0]
	}
	if q.ExpectEmpty {
		res.HitAt5 = len(ids) == 0
		if res.HitAt5 {
			res.RR = 1
		}
		return res
	}
	expected := make(map[string]struct{}, len(q.ExpectedIDs))
	for _, id := range q.ExpectedIDs {
		expected[strings.ToLower(id)] = struct{}{}
	}
	for i, id := range ids {
		if _, ok := expected[strings.ToLower(id)]; ok {
			res.HitAt5 = true
			res.RR = 1.0 / float64(i+1)
			break
		}
	}
	return res
}
