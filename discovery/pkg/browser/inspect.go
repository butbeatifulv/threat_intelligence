package browser

import (
	"encoding/json"
	"strconv"
)

// InspectRequest configures browser inspection.
type InspectRequest struct {
	URL         string
	WaitTime    int
	Headless    bool
	Screenshot  bool
	ActiveTests bool
}

// InspectResult is the normalized browser inspect response.
type InspectResult struct {
	Success          bool             `json:"success"`
	Forms            []map[string]any `json:"forms,omitempty"`
	PageInfo         map[string]any   `json:"page_info,omitempty"`
	SecurityAnalysis map[string]any   `json:"security_analysis,omitempty"`
	Technologies     []string         `json:"technologies,omitempty"`
	Screenshot       string           `json:"screenshot,omitempty"`
	Timestamp        string           `json:"timestamp,omitempty"`
	Error            string           `json:"error,omitempty"`
}

// InspectFromParams maps catalog/MCP parameters to InspectRequest.
func InspectFromParams(target string, params map[string]string) InspectRequest {
	url := target
	if params != nil {
		if u := params["url"]; u != "" {
			url = u
		}
	}
	wait := 5
	if params != nil {
		if w := params["wait_time"]; w != "" {
			if n, err := strconv.Atoi(w); err == nil {
				wait = n
			}
		}
	}
	headless := params == nil || (params["headless"] != "false" && params["headless"] != "False")

	active := params != nil && (params["active_tests"] == "true" || params["active_tests"] == "True")

	return InspectRequest{
		URL:         url,
		WaitTime:    wait,
		Headless:    headless,
		Screenshot:  true,
		ActiveTests: active,
	}
}

func normalizeInspectResult(r *InspectResult) {
	if r == nil {
		return
	}
	if len(r.Forms) == 0 && r.PageInfo != nil {
		r.Forms = formsFromAny(r.PageInfo["forms"])
	}
}

func formsFromAny(raw any) []map[string]any {
	arr, ok := raw.([]any)
	if !ok || len(arr) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(arr))
	for _, it := range arr {
		m, ok := it.(map[string]any)
		if ok {
			out = append(out, m)
		}
	}
	return out
}

// ToJSON returns inspect result as JSON string for tool output.
func (r InspectResult) ToJSON() string {
	b, _ := json.Marshal(r)
	return string(b)
}
