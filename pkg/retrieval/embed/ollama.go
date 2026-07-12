package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Ollama embeds via local Ollama API.
type Ollama struct {
	BaseURL string
	Model   string
	Client  *http.Client
}

func (o *Ollama) client() *http.Client {
	if o.Client != nil {
		return o.Client
	}
	return &http.Client{Timeout: 120 * time.Second}
}

func (o *Ollama) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	url := strings.TrimRight(o.BaseURL, "/") + "/api/embeddings"
	body := map[string]any{"model": o.model(), "prompt": texts[0]}
	if len(texts) > 1 {
		body["prompt"] = texts
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := o.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama embed: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	var out struct {
		Embedding  []float32   `json:"embedding"`
		Embeddings [][]float32 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Embeddings) > 0 {
		return out.Embeddings, nil
	}
	if len(out.Embedding) > 0 {
		return [][]float32{out.Embedding}, nil
	}
	return nil, fmt.Errorf("ollama embed: empty response")
}

func (o *Ollama) Dimensions() int { return 768 }

func (o *Ollama) ModelID() string { return o.model() }

func (o *Ollama) model() string {
	if m := strings.TrimSpace(o.Model); m != "" {
		return m
	}
	return "nomic-embed-text"
}
