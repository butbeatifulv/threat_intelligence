package usecase

import (
	"context"

	"github.com/butbeautifulv/veil/knowledge/connector/query"
)

// DefaultTargetGraphCategories are category IDs searched for target decisions.
var DefaultTargetGraphCategories = []string{"ti", "vuln", "engage"}

// TargetGraphOpts configures TargetGraph aggregation.
type TargetGraphOpts struct {
	Categories           []string
	IncludeEngageContext bool
	SearchQuery          string
}

// TargetGraphState is the unified read model for a target host across graph categories.
type TargetGraphState struct {
	Target        string                     `json:"target"`
	Host          string                     `json:"host"`
	Hits          map[string][]query.Node    `json:"hits,omitempty"`
	EngageContext *query.EngageTargetContext `json:"engage_context,omitempty"`
	EngageFound   bool                       `json:"engage_found"`
}

// TargetGraph loads category search hits and optional engage subgraph for a target host.
func (u *ReadUsecase) TargetGraph(ctx context.Context, target string, opts TargetGraphOpts) (TargetGraphState, error) {
	state := TargetGraphState{
		Target: target,
		Host:   query.NormalizeEngageHost(target),
		Hits:   map[string][]query.Node{},
	}
	q := opts.SearchQuery
	if q == "" {
		q = state.Host
	}
	cats := opts.Categories
	if len(cats) == 0 {
		cats = DefaultTargetGraphCategories
	}
	if q != "" {
		for _, cat := range cats {
			nodes, err := u.SearchInCategory(ctx, cat, q, "", 50)
			if err != nil {
				return state, err
			}
			if len(nodes) > 0 {
				state.Hits[cat] = nodes
			}
		}
	}
	if opts.IncludeEngageContext && state.Host != "" {
		ectx, err := u.EngageTargetContext(ctx, state.Host)
		if err != nil {
			return state, err
		}
		if ectx != nil {
			state.EngageContext = ectx
			state.EngageFound = ectx.Target != nil
		}
	}
	return state, nil
}
