package embed

import (
	"context"
	"fmt"
)

// Provider embeds text batches.
type Provider interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	Dimensions() int
	ModelID() string
}

// NoopProvider fails embedding (CI / keyword-only mode).
type NoopProvider struct{}

func (NoopProvider) Embed(context.Context, []string) ([][]float32, error) {
	return nil, fmt.Errorf("embed: noop provider")
}

func (NoopProvider) Dimensions() int { return 0 }

func (NoopProvider) ModelID() string { return "noop" }

// NewFromConfig selects provider by name.
func NewFromConfig(provider, url, model, apiKey string) Provider {
	switch provider {
	case "ollama":
		return &Ollama{BaseURL: url, Model: model}
	case "openai", "voyage":
		return &OpenAI{BaseURL: url, Model: model, APIKey: apiKey}
	default:
		return NoopProvider{}
	}
}
