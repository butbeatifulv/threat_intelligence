package retrieval

import (
	"fmt"
	"strings"
	"sync"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search"
)

// BleveIndex wraps a Bleve keyword index.
type BleveIndex struct {
	idx *bleve.Index
}

var (
	bleveOpen   func(path string) (*BleveIndex, error) = openBleve
	bleveOpenMu sync.Mutex
)

func openBleve(path string) (*BleveIndex, error) {
	idx, err := bleve.Open(path)
	if err != nil {
		return nil, fmt.Errorf("bleve open %s: %w", path, err)
	}
	return &BleveIndex{idx: &idx}, nil
}

// OpenBleve opens or reuses a Bleve index at path.
func OpenBleve(path string) (*BleveIndex, error) {
	bleveOpenMu.Lock()
	defer bleveOpenMu.Unlock()
	return bleveOpen(path)
}

// Close closes the underlying index.
func (b *BleveIndex) Close() error {
	if b == nil || b.idx == nil {
		return nil
	}
	idx := *b.idx
	b.idx = nil
	return idx.Close()
}

// Search runs BM25 search over chunk text.
func (b *BleveIndex) Search(q string, subdomain string, limit int) ([]ChunkHit, error) {
	if b == nil || b.idx == nil {
		return nil, fmt.Errorf("bleve: closed index")
	}
	if limit <= 0 {
		limit = DefaultSearchLimit
	}
	qq := bleve.NewMatchQuery(strings.TrimSpace(q))
	qq.SetField("text")
	req := bleve.NewSearchRequest(qq)
	req.Size = limit * 4
	if subdomain != "" {
		subQ := bleve.NewTermQuery(strings.ToLower(subdomain))
		subQ.SetField("subdomain")
		req.Query = bleve.NewConjunctionQuery(qq, subQ)
	}
	req.Fields = []string{"skill_id", "subdomain", "section_title", "text", "attack_ids"}
	res, err := (*b.idx).Search(req)
	if err != nil {
		return nil, err
	}
	out := make([]ChunkHit, 0, len(res.Hits))
	for _, hit := range res.Hits {
		out = append(out, chunkFromHit(hit))
	}
	return out, nil
}

func chunkFromHit(hit *search.DocumentMatch) ChunkHit {
	get := func(k string) string {
		if hit.Fields == nil {
			return ""
		}
		if v, ok := hit.Fields[k]; ok {
			return fmt.Sprint(v)
		}
		return ""
	}
	text := get("text")
	snippet := text
	if len(snippet) > 200 {
		snippet = snippet[:200]
	}
	return ChunkHit{
		SkillID:      get("skill_id"),
		Subdomain:    get("subdomain"),
		SectionTitle: get("section_title"),
		Text:         text,
		Score:        hit.Score,
		MatchType:    MatchKeyword,
	}
}

// MatchAllQuery for tests — kept for future bleve query tuning.
