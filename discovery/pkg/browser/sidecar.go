package browser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Sidecar calls the Playwright browser-agent HTTP API.
type Sidecar struct {
	BaseURL string
	Client  *http.Client
}

// NewSidecarFromEnv returns a sidecar client when DISCOVERY_BROWSER_SIDECAR_URL or
// ENGAGE_BROWSER_SIDECAR_URL is set (legacy engage compose name).
func NewSidecarFromEnv() *Sidecar {
	base := strings.TrimSpace(os.Getenv("DISCOVERY_BROWSER_SIDECAR_URL"))
	if base == "" {
		base = strings.TrimSpace(os.Getenv("ENGAGE_BROWSER_SIDECAR_URL"))
	}
	if base == "" {
		return nil
	}
	return &Sidecar{
		BaseURL: strings.TrimRight(base, "/"),
		Client:  &http.Client{Timeout: 5 * time.Minute},
	}
}

func (s *Sidecar) Enabled() bool {
	return s != nil && s.BaseURL != ""
}

// Inspect runs POST /inspect on the Playwright sidecar.
func (s *Sidecar) Inspect(ctx context.Context, req InspectRequest) InspectResult {
	if s == nil || !s.Enabled() {
		return InspectResult{Success: false, Error: "browser sidecar not configured (DISCOVERY_BROWSER_SIDECAR_URL)"}
	}
	url := strings.TrimSpace(req.URL)
	if url == "" {
		return InspectResult{Success: false, Error: "url required"}
	}
	payload := map[string]any{
		"url":          url,
		"target":       url,
		"wait_time":    req.WaitTime,
		"headless":     req.Headless,
		"screenshot":   req.Screenshot,
		"active_tests": req.ActiveTests,
		"inspect":      true,
	}
	body, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.BaseURL+"/inspect", bytes.NewReader(body))
	if err != nil {
		return InspectResult{Success: false, Error: err.Error()}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := s.Client.Do(httpReq)
	if err != nil {
		return InspectResult{Success: false, Error: err.Error()}
	}
	defer resp.Body.Close()
	var out InspectResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return InspectResult{Success: false, Error: fmt.Sprintf("decode: %v", err)}
	}
	normalizeInspectResult(&out)
	if resp.StatusCode >= 400 && out.Error == "" {
		out.Success = false
		out.Error = fmt.Sprintf("browser sidecar http %d", resp.StatusCode)
	}
	return out
}
