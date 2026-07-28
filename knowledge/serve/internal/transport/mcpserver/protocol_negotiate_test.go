package mcpserver

import (
	"encoding/json"
	"testing"
)

func TestNegotiateProtocol(t *testing.T) {
	tests := []struct {
		name   string
		params string
		want   string
	}{
		{"empty", "", defaultProtocol},
		{"2024-11-05", `{"protocolVersion":"2024-11-05"}`, protocol20241105},
		{"2025-03-26", `{"protocolVersion":"2025-03-26"}`, protocol20250326},
		{"unknown", `{"protocolVersion":"2099-01-01"}`, defaultProtocol},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw json.RawMessage
			if tt.params != "" {
				raw = json.RawMessage(tt.params)
			}
			if got := negotiateProtocol(raw); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestProcessMessage_initialize_negotiates2024(t *testing.T) {
	srv := NewServer(nil, nil, nil, nil, nil, nil)
	msg := rpcMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion":"2024-11-05","capabilities":{}}`),
	}
	resp, notif, err := srv.ProcessMessage(t.Context(), msg, false)
	if err != nil || notif || resp == nil {
		t.Fatalf("resp=%+v notif=%v err=%v", resp, notif, err)
	}
	pv, _ := resp.Result.(map[string]any)["protocolVersion"].(string)
	if pv != protocol20241105 {
		t.Fatalf("protocol %q", pv)
	}
}

func TestProcessMessage_initialize_negotiates2025(t *testing.T) {
	srv := NewServer(nil, nil, nil, nil, nil, nil)
	msg := rpcMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion":"2025-03-26","capabilities":{}}`),
	}
	resp, _, err := srv.ProcessMessage(t.Context(), msg, true)
	if err != nil || resp == nil {
		t.Fatal(err)
	}
	pv, _ := resp.Result.(map[string]any)["protocolVersion"].(string)
	if pv != protocol20250326 {
		t.Fatalf("protocol %q", pv)
	}
}
