package mcp

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestListToolsPayload(t *testing.T) {
	tools := []ToolDescriptor{
		{
			Name:        "scan",
			Description: "run scan",
			InputSchema: map[string]any{"type": "object"},
		},
	}
	got := ListToolsPayload(tools)
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Tools []struct {
			Name        string         `json:"name"`
			Description string         `json:"description"`
			InputSchema map[string]any `json:"inputSchema"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Tools) != 1 || decoded.Tools[0].Name != "scan" {
		t.Fatalf("tools: %+v", decoded.Tools)
	}
	if decoded.Tools[0].Description != "run scan" {
		t.Fatalf("description: %q", decoded.Tools[0].Description)
	}
	if decoded.Tools[0].InputSchema["type"] != "object" {
		t.Fatalf("inputSchema: %+v", decoded.Tools[0].InputSchema)
	}
}

func TestListToolsPayload_empty(t *testing.T) {
	got := ListToolsPayload(nil)
	tools, ok := got["tools"].([]map[string]any)
	if !ok {
		t.Fatalf("tools key: %T", got["tools"])
	}
	if len(tools) != 0 {
		t.Fatalf("expected empty slice, got %v", tools)
	}
}

func TestParseToolCallParams_ok(t *testing.T) {
	params := json.RawMessage(`{"name":"scan","arguments":{"target":"10.0.0.1"}}`)
	got, err := ParseToolCallParams(params)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "scan" || got.Arguments["target"] != "10.0.0.1" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseToolCallParams_invalidJSON(t *testing.T) {
	_, err := ParseToolCallParams(json.RawMessage(`{not json`))
	var re *RPCError
	if !errors.As(err, &re) || re.Code != CodeInvalidParams {
		t.Fatalf("got %v", err)
	}
}

func TestParseToolCallParams_missingName(t *testing.T) {
	_, err := ParseToolCallParams(json.RawMessage(`{"arguments":{}}`))
	var re *RPCError
	if !errors.As(err, &re) || re.Code != CodeInvalidParams || re.Message != "tool name required" {
		t.Fatalf("got %v", err)
	}
}
