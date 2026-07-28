package retrieval

import (
	"fmt"
	"sort"
	"strings"
)

// Engine orchestrates hybrid retrieval.
type Engine struct {
	cfg   Config
	bleve *BleveIndex
	qdrant *QdrantClient
	embed  Embedder
	rerank *RerankClient
}

// Embedder embeds text for vector search.
type Embedder interface {
	EmbedQuery(text string) ([]float32, error)
	ModelID() string
}

// NewEngine opens indexes per config.
func NewEngine(cfg Config) (*Engine, error) {
	e := &Engine{cfg: cfg}
	if cfg.BleveIndexPath != "" {
		b, err := OpenBleve(cfg.BleveIndexPath)
		if err != nil {
			return nil, err
		}
		e.bleve = b
	}
	if cfg.QdrantURL != "" && cfg.SearchMode != ModeKeyword {
		e.qdrant = NewQdrantClient(cfg.QdrantURL, cfg.QdrantCollection)
	}
	return e, nil
}

// Close releases resources.
func (e *Engine) Close() error {
	if e == nil {
		return nil
	}
	return e.bleve.Close()
}

// Search runs retrieval per configured mode.
func (e *Engine) Search(opts SearchOpts) ([]SearchResult, error) {
	if e == nil {
		return nil, fmt.Errorf("retrieval: nil engine")
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = DefaultSearchLimit
	}
	mode := opts.Mode
	if mode == "" {
		mode = e.cfg.SearchMode
	}
	fetchK := e.cfg.FetchK
	if fetchK < limit {
		fetchK = limit * 3
	}

	var lists [][]RankedHit
	switch mode {
	case ModeKeyword:
		hits, err := e.keywordHits(opts.Query, opts.Subdomain, fetchK)
		if err != nil {
			return nil, err
		}
		lists = append(lists, rankedFromChunks(hits))
	case ModeVector:
		hits, err := e.vectorHits(opts.Query, opts.Subdomain, fetchK)
		if err != nil {
			return nil, err
		}
		lists = append(lists, rankedFromChunks(hits))
	default:
		var kw, vec []ChunkHit
		var err, vErr error
		if e.bleve != nil {
			kw, err = e.keywordHits(opts.Query, opts.Subdomain, fetchK)
		}
		if e.qdrant != nil && e.embed != nil {
			vec, vErr = e.vectorHits(opts.Query, opts.Subdomain, fetchK)
		}
		if len(kw) == 0 && len(vec) == 0 {
			if err != nil {
				return nil, err
			}
			if vErr != nil {
				return nil, vErr
			}
			return nil, nil
		}
		if len(kw) > 0 {
			lists = append(lists, rankedFromChunks(kw))
		}
		if len(vec) > 0 {
			lists = append(lists, rankedFromChunks(vec))
		}
	}

	var ranked []RankedHit
	if len(lists) == 1 {
		ranked = lists[0]
	} else if len(lists) > 1 {
		ranked = RRF(lists, e.cfg.RRFK, nil)
	}
	ranked = ApplyExactBoost(opts.Query, ranked)
	results := dedupeResults(ranked, limit)
	if e.rerank != nil && e.cfg.RerankEnabled && len(results) > limit {
		var err error
		results, err = e.rerank.Rerank(opts.Query, results, limit)
		if err != nil {
			// fallback to fusion order
			if len(results) > limit {
				results = results[:limit]
			}
		}
	}
	return results, nil
}

func (e *Engine) keywordHits(q, subdomain string, limit int) ([]ChunkHit, error) {
	if e.bleve == nil {
		return nil, fmt.Errorf("bleve index not configured")
	}
	return e.bleve.Search(q, subdomain, limit)
}

func (e *Engine) vectorHits(q, subdomain string, limit int) ([]ChunkHit, error) {
	if e.qdrant == nil || e.embed == nil {
		return nil, nil
	}
	vec, err := e.embed.EmbedQuery(q)
	if err != nil {
		return nil, err
	}
	return e.qdrant.Search(vec, subdomain, limit)
}

func rankedFromChunks(hits []ChunkHit) []RankedHit {
	out := make([]RankedHit, 0, len(hits))
	for _, h := range hits {
		out = append(out, RankedHit{SkillID: h.SkillID, Score: h.Score, Hit: h})
	}
	return out
}

func dedupeResults(ranked []RankedHit, limit int) []SearchResult {
	bySkill := map[string]SearchResult{}
	order := []string{}
	for _, r := range ranked {
		sid := strings.ToLower(r.SkillID)
		if sid == "" {
			continue
		}
		snippet := r.Hit.Text
		if len(snippet) > 200 {
			snippet = snippet[:200]
		}
		mt := r.Hit.MatchType
		if mt == "" {
			mt = MatchKeyword
		}
		prev, ok := bySkill[sid]
		if !ok || r.Score > prev.Score {
			bySkill[sid] = SearchResult{
				SkillID:   r.SkillID,
				Score:     r.Score,
				Snippet:   snippet,
				MatchType: mt,
			}
			if !ok {
				order = append(order, sid)
			}
		}
	}
	sort.Slice(order, func(i, j int) bool {
		return bySkill[order[i]].Score > bySkill[order[j]].Score
	})
	out := make([]SearchResult, 0, limit)
	for _, sid := range order {
		out = append(out, bySkill[sid])
		if len(out) >= limit {
			break
		}
	}
	return out
}

// SetEmbedder wires vector search (optional).
func (e *Engine) SetEmbedder(p Embedder) {
	if e != nil {
		e.embed = p
	}
}

// SetReranker wires cross-encoder reranking (optional).
func (e *Engine) SetReranker(r *RerankClient) {
	if e != nil {
		e.rerank = r
	}
}

// SearchIDs returns skill ids for eval harness.
func (e *Engine) SearchIDs(query, subdomain string, limit int) ([]string, error) {
	res, err := e.Search(SearchOpts{Query: query, Subdomain: subdomain, Limit: limit, Mode: e.cfg.SearchMode})
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(res))
	for _, r := range res {
		ids = append(ids, r.SkillID)
	}
	return ids, nil
}
