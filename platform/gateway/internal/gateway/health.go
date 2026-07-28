package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/butbeautifulv/veil/pkg/api"
	"github.com/butbeautifulv/veil/platform/gateway/internal/config"
)

type upstreamCheck struct {
	Name string
	URL  string
}

func apiHealthChecks(cfg config.Config) []upstreamCheck {
	return []upstreamCheck{
		{Name: "graph_api", URL: cfg.GraphAPIURL.String() + "/health"},
		{Name: "engage_api", URL: cfg.EngageAPIURL.String() + "/health"},
	}
}

// UpstreamsHealthy reports whether all configured API upstream health endpoints respond ok.
func UpstreamsHealthy(ctx context.Context, cfg config.Config) bool {
	results := probeUpstreams(ctx, apiHealthChecks(cfg))
	for _, res := range results {
		if !res.OK {
			return false
		}
	}
	return true
}

func registerCompositeHealth(mux *http.ServeMux, cfg config.Config) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		results := probeUpstreams(ctx, apiHealthChecks(cfg))
		ok := true
		for _, res := range results {
			if !res.OK {
				ok = false
				break
			}
		}

		body := map[string]any{
			"ok":        ok,
			"service":   config.ServiceName(),
			"upstreams": results,
		}
		status := http.StatusOK
		if !ok {
			status = http.StatusServiceUnavailable
		}
		api.WriteJSON(w, status, body)
	})
}

type upstreamResult struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	OK     bool   `json:"ok"`
	Status int    `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

func probeUpstreams(ctx context.Context, checks []upstreamCheck) []upstreamResult {
	out := make([]upstreamResult, len(checks))
	var wg sync.WaitGroup
	client := &http.Client{Timeout: 4 * time.Second}
	for i, chk := range checks {
		wg.Add(1)
		go func(i int, chk upstreamCheck) {
			defer wg.Done()
			res := upstreamResult{Name: chk.Name, URL: chk.URL}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, chk.URL, nil)
			if err != nil {
				res.Error = err.Error()
				out[i] = res
				return
			}
			resp, err := client.Do(req)
			if err != nil {
				res.Error = err.Error()
				out[i] = res
				return
			}
			defer resp.Body.Close()
			res.Status = resp.StatusCode
			if resp.StatusCode == http.StatusOK {
				var body map[string]any
				if json.NewDecoder(resp.Body).Decode(&body) == nil {
					if v, _ := body["ok"].(bool); v {
						res.OK = true
					}
				}
			}
			if !res.OK && res.Error == "" {
				res.Error = "upstream health not ok"
			}
			out[i] = res
		}(i, chk)
	}
	wg.Wait()
	return out
}
