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

// QdrantClient queries Qdrant REST API.
type QdrantClient struct {
	BaseURL    string
	Collection string
	Client     *http.Client
}

// NewQdrantClient creates a Qdrant HTTP client.
func NewQdrantClient(baseURL, collection string) *QdrantClient {
	return &QdrantClient{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Collection: collection,
		Client:     &http.Client{Timeout: 30 * time.Second},
	}
}

// Search finds nearest chunk vectors.
func (q *QdrantClient) Search(vector []float32, subdomain string, limit int) ([]ChunkHit, error) {
	if q == nil {
		return nil, fmt.Errorf("qdrant: nil client")
	}
	if len(vector) == 0 {
		return nil, fmt.Errorf("qdrant: empty vector")
	}
	if limit <= 0 {
		limit = DefaultSearchLimit
	}
	filter := any(nil)
	if subdomain != "" {
		filter = map[string]any{
			"must": []map[string]any{
				{"key": "subdomain", "match": map[string]any{"value": subdomain}},
			},
		}
	}
	body, _ := json.Marshal(map[string]any{
		"vector":       vector,
		"limit":        limit,
		"with_payload": true,
		"filter":       filter,
	})
	url := fmt.Sprintf("%s/collections/%s/points/search", q.BaseURL, q.Collection)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := q.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("qdrant search: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	var out struct {
		Result []struct {
			Score   float64        `json:"score"`
			Payload map[string]any `json:"payload"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	hits := make([]ChunkHit, 0, len(out.Result))
	for _, r := range out.Result {
		p := r.Payload
		text := strPayload(p, "text")
		hits = append(hits, ChunkHit{
			SkillID:      strPayload(p, "skill_id"),
			Subdomain:    strPayload(p, "subdomain"),
			SectionTitle: strPayload(p, "section_title"),
			Text:         text,
			Score:        r.Score,
			MatchType:    MatchVector,
		})
	}
	return hits, nil
}

func (q *QdrantClient) client() *http.Client {
	if q.Client != nil {
		return q.Client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func strPayload(p map[string]any, key string) string {
	if p == nil {
		return ""
	}
	if v, ok := p[key]; ok {
		return fmt.Sprint(v)
	}
	return ""
}
