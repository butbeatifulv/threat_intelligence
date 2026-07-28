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

// OpenAI uses OpenAI-compatible embeddings API (OpenAI, Voyage, etc.).
type OpenAI struct {
	BaseURL string
	Model   string
	APIKey  string
	Client  *http.Client
}

func (o *OpenAI) client() *http.Client {
	if o.Client != nil {
		return o.Client
	}
	return &http.Client{Timeout: 120 * time.Second}
}

func (o *OpenAI) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	base := strings.TrimRight(o.BaseURL, "/")
	url := base
	if !strings.HasSuffix(base, "/embeddings") {
		url = base + "/v1/embeddings"
	}
	body, _ := json.Marshal(map[string]any{
		"model": o.model(),
		"input": texts,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if k := strings.TrimSpace(o.APIKey); k != "" {
		req.Header.Set("Authorization", "Bearer "+k)
	}
	resp, err := o.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai embed: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	var out struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	vecs := make([][]float32, len(texts))
	for _, item := range out.Data {
		if item.Index >= 0 && item.Index < len(vecs) {
			vecs[item.Index] = item.Embedding
		}
	}
	return vecs, nil
}

func (o *OpenAI) Dimensions() int { return 1536 }

func (o *OpenAI) ModelID() string { return o.model() }

func (o *OpenAI) model() string {
	if m := strings.TrimSpace(o.Model); m != "" {
		return m
	}
	return "text-embedding-3-small"
}
