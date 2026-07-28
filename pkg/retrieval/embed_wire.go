package retrieval

import (
	"context"

	"github.com/butbeautifulv/veil/pkg/retrieval/embed"
)

type embedAdapter struct {
	p embed.Provider
}

func (a embedAdapter) EmbedQuery(text string) ([]float32, error) {
	vecs, err := a.p.Embed(context.Background(), []string{text})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, nil
	}
	return vecs[0], nil
}

func (a embedAdapter) ModelID() string { return a.p.ModelID() }

// WireEmbedder attaches configured embed provider to engine.
func WireEmbedder(e *Engine, cfg Config) {
	if e == nil {
		return
	}
	p := embed.NewFromConfig(cfg.EmbedProvider, cfg.EmbedURL, cfg.EmbedModel, cfg.EmbedAPIKey)
	e.SetEmbedder(embedAdapter{p: p})
}
