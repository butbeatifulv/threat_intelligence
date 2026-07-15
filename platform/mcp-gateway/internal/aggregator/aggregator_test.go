package aggregator_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/butbeautifulv/veil/platform/mcp-gateway/internal/aggregator"
	"github.com/butbeautifulv/veil/platform/mcp-gateway/internal/backend"
	"github.com/butbeautifulv/veil/pkg/mcp"
)

func mockMCPBackend(t *testing.T, tools []map[string]any, calls map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var msg mcp.Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch msg.Method {
		case "tools/list":
			_ = json.NewEncoder(w).Encode(mcp.Message{
				JSONRPC: "2.0",
				ID:      msg.ID,
				Result:  map[string]any{"tools": toAnySlice(tools)},
			})
		case "tools/call":
			var p struct {
				Name string `json:"name"`
			}
			_ = json.Unmarshal(msg.Params, &p)
			res, ok := calls[p.Name]
			if !ok {
				_ = json.NewEncoder(w).Encode(mcp.Message{
					JSONRPC: "2.0",
					ID:      msg.ID,
					Error:   &mcp.RPCError{Code: mcp.CodeInvalidParams, Message: "unknown tool"},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(mcp.Message{
				JSONRPC: "2.0",
				ID:      msg.ID,
				Result:  res,
			})
		default:
			http.Error(w, "unexpected method", http.StatusBadRequest)
		}
	}))
}

func toAnySlice(in []map[string]any) []any {
	out := make([]any, len(in))
	for i, m := range in {
		out[i] = m
	}
	return out
}

func TestAggregator_listTools_mergesWithPrefixes(t *testing.T) {
	graphSrv := mockMCPBackend(t,
		[]map[string]any{{"name": "ti_health", "description": "graph health"}},
		nil,
	)
	defer graphSrv.Close()
	engageSrv := mockMCPBackend(t,
		[]map[string]any{{"name": "nmap", "description": "scan"}},
		nil,
	)
	defer engageSrv.Close()

	agg := aggregator.New(
		&backend.Client{Name: "graph", URL: graphSrv.URL, HTTP: graphSrv.Client()},
		&backend.Client{Name: "engage", URL: engageSrv.URL, HTTP: engageSrv.Client()},
		nil,
		nil,
	)

	raw, _, err := agg.ProcessMessage(context.Background(), mcp.Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	if raw == nil {
		t.Fatal("nil response")
	}
	result, ok := raw.Result.(map[string]any)
	if !ok {
		t.Fatalf("result type %T", raw.Result)
	}
	tools, ok := result["tools"].([]any)
	if !ok || len(tools) != 2 {
		t.Fatalf("tools %#v", result["tools"])
	}
	names := map[string]bool{}
	for _, item := range tools {
		m := item.(map[string]any)
		names[m["name"].(string)] = true
		ann := m["annotations"].(map[string]any)
		if ann["x-veil-backend"] == "" {
			t.Fatalf("missing backend metadata on %v", m["name"])
		}
	}
	if !names["graph_ti_health"] || !names["engage_nmap"] {
		t.Fatalf("names %v", names)
	}
}

func TestAggregator_listTools_graphOnlyWhenEngageDown(t *testing.T) {
	graphSrv := mockMCPBackend(t,
		[]map[string]any{{"name": "ti_health", "description": "graph health"}},
		nil,
	)
	defer graphSrv.Close()
	engageSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "engage down", http.StatusServiceUnavailable)
	}))
	defer engageSrv.Close()

	agg := aggregator.New(
		&backend.Client{Name: "graph", URL: graphSrv.URL, HTTP: graphSrv.Client()},
		&backend.Client{Name: "engage", URL: engageSrv.URL, HTTP: engageSrv.Client()},
		nil,
		nil,
	)

	raw, _, err := agg.ProcessMessage(context.Background(), mcp.Message{
		JSONRPC: "2.0",
		ID:      9,
		Method:  "tools/list",
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	result := raw.Result.(map[string]any)
	tools := result["tools"].([]any)
	if len(tools) != 1 {
		t.Fatalf("want 1 graph tool, got %#v", tools)
	}
	if tools[0].(map[string]any)["name"] != "graph_ti_health" {
		t.Fatalf("tool %#v", tools[0])
	}
}

func TestAggregator_callTool_routesByPrefix(t *testing.T) {
	graphSrv := mockMCPBackend(t, nil, map[string]any{
		"ti_health": map[string]any{"content": []any{map[string]any{"type": "text", "text": "ok"}}},
	})
	defer graphSrv.Close()
	engageSrv := mockMCPBackend(t, nil, map[string]any{
		"nmap": map[string]any{"content": []any{map[string]any{"type": "text", "text": "scan"}}},
	})
	defer engageSrv.Close()

	agg := aggregator.New(
		&backend.Client{Name: "graph", URL: graphSrv.URL, HTTP: graphSrv.Client()},
		&backend.Client{Name: "engage", URL: engageSrv.URL, HTTP: engageSrv.Client()},
		nil,
		nil,
	)

	params, _ := json.Marshal(map[string]any{
		"name":      "graph_ti_health",
		"arguments": map[string]any{},
	})
	raw, _, err := agg.ProcessMessage(context.Background(), mcp.Message{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params:  params,
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	res := raw.Result.(map[string]any)
	content := res["content"].([]any)
	if content[0].(map[string]any)["text"] != "ok" {
		t.Fatalf("graph result %#v", res)
	}

	params, _ = json.Marshal(map[string]any{
		"name":      "engage_nmap",
		"arguments": map[string]any{"target": "127.0.0.1"},
	})
	raw, _, err = agg.ProcessMessage(context.Background(), mcp.Message{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params:  params,
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	res = raw.Result.(map[string]any)
	content = res["content"].([]any)
	if content[0].(map[string]any)["text"] != "scan" {
		t.Fatalf("engage result %#v", res)
	}
}

func TestHTTP_initialize(t *testing.T) {
	graphSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer graphSrv.Close()
	engageSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer engageSrv.Close()

	agg := aggregator.New(
		&backend.Client{Name: "graph", URL: graphSrv.URL, HTTP: graphSrv.Client()},
		&backend.Client{Name: "engage", URL: engageSrv.URL, HTTP: engageSrv.Client()},
		nil,
		nil,
	)
	h := mcp.HTTPHandler(agg, mcp.HTTPConfig{Path: "/mcp", Service: aggregator.ServerName})
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d %s", rr.Code, rr.Body.String())
	}
	var msg mcp.Message
	if err := json.Unmarshal(rr.Body.Bytes(), &msg); err != nil {
		t.Fatal(err)
	}
	info := msg.Result.(map[string]any)["serverInfo"].(map[string]any)
	if info["name"] != aggregator.ServerName {
		t.Fatalf("server %v", info)
	}
}
