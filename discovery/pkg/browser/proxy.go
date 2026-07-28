package browser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ExecResult is the catalog browser tool stdout contract.
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

// CatalogProxy runs browser catalog tools via a discovery browser HTTP service.
type CatalogProxy struct {
	BaseURL string
	Client  *http.Client
}

// NewCatalogProxy targets the discovery browser service (inspect + exec).
func NewCatalogProxy(baseURL string) *CatalogProxy {
	return &CatalogProxy{
		BaseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		Client:  &http.Client{Timeout: 5 * time.Minute},
	}
}

func (b *CatalogProxy) Enabled() bool {
	return b != nil && b.BaseURL != ""
}

// IsBrowserBinary reports catalog placeholder binary "browser".
func IsBrowserBinary(name string) bool {
	return strings.EqualFold(strings.TrimSpace(name), "browser")
}

// Inspect posts to /inspect on the discovery browser service.
func (b *CatalogProxy) Inspect(ctx context.Context, req InspectRequest) ExecResult {
	body, _ := json.Marshal(map[string]any{
		"url":           req.URL,
		"target":        req.URL,
		"wait_time":     req.WaitTime,
		"headless":      req.Headless,
		"screenshot":    req.Screenshot,
		"active_tests":  req.ActiveTests,
		"inspect":       true,
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, b.BaseURL+"/inspect", bytes.NewReader(body))
	if err != nil {
		return ExecResult{ExitCode: -1, Err: err}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := b.Client.Do(httpReq)
	if err != nil {
		return ExecResult{ExitCode: -1, Err: err}
	}
	defer resp.Body.Close()
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return ExecResult{ExitCode: -1, Err: fmt.Errorf("browser service: %w", err)}
	}
	if raw["success"] == false {
		errMsg, _ := raw["error"].(string)
		return ExecResult{ExitCode: 1, Err: fmt.Errorf("%s", errMsg), Stderr: errMsg}
	}
	out, _ := json.Marshal(raw)
	return ExecResult{Stdout: string(out) + "\n", ExitCode: 0}
}

// Exec runs catalog browser via POST /exec.
func (b *CatalogProxy) Exec(ctx context.Context, target string, args []string) ExecResult {
	body, _ := json.Marshal(map[string]any{
		"url":    target,
		"target": target,
		"args":   args,
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, b.BaseURL+"/exec", bytes.NewReader(body))
	if err != nil {
		return ExecResult{ExitCode: -1, Err: err}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := b.Client.Do(httpReq)
	if err != nil {
		return ExecResult{ExitCode: -1, Err: err}
	}
	defer resp.Body.Close()
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return ExecResult{ExitCode: -1, Err: err}
	}
	exitCode := 0
	if v, ok := raw["exit_code"].(float64); ok {
		exitCode = int(v)
	}
	stdout, _ := raw["stdout"].(string)
	stderr, _ := raw["stderr"].(string)
	errMsg, _ := raw["error"].(string)
	var runErr error
	if exitCode != 0 {
		runErr = fmt.Errorf("%s", errMsg)
		if runErr.Error() == "" {
			runErr = fmt.Errorf("browser exec exit %d", exitCode)
		}
	}
	return ExecResult{Stdout: stdout, Stderr: stderr, ExitCode: exitCode, Err: runErr}
}
