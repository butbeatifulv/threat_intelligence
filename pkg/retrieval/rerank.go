package retrieval

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RerankClient calls optional cross-encoder rerank sidecar.
type RerankClient struct {
	BaseURL string
	Client  *http.Client
}

// NewRerankClient creates rerank HTTP client.
func NewRerankClient(baseURL string) *RerankClient {
	return &RerankClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Client:  &http.Client{Timeout: 60 * time.Second},
	}
}

// Rerank reorders results by query relevance.
func (r *RerankClient) Rerank(query string, results []SearchResult, limit int) ([]SearchResult, error) {
	if r == nil || r.BaseURL == "" {
		return results, fmt.Errorf("rerank: not configured")
	}
	docs := make([]string, len(results))
	for i, res := range results {
		docs[i] = res.Snippet
	}
	body, _ := json.Marshal(map[string]any{
		"query": query,
		"docs":  docs,
		"top_k": limit,
	})
	req, err := http.NewRequest(http.MethodPost, r.BaseURL+"/rerank", bytes.NewReader(body))
	if err != nil {
		return results, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.client().Do(req)
	if err != nil {
		return results, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return results, fmt.Errorf("rerank: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	var out struct {
		Indices []int `json:"indices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return results, err
	}
	reranked := make([]SearchResult, 0, limit)
	for _, idx := range out.Indices {
		if idx < 0 || idx >= len(results) {
			continue
		}
		reranked = append(reranked, results[idx])
		if len(reranked) >= limit {
			break
		}
	}
	return reranked, nil
}

func (r *RerankClient) client() *http.Client {
	if r.Client != nil {
		return r.Client
	}
	return &http.Client{Timeout: 60 * time.Second}
}
