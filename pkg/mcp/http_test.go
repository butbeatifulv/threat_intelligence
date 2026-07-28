package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubProc struct{}

func (stubProc) ProcessMessage(ctx context.Context, msg Message, httpTransport bool) (*Message, bool, error) {
	if msg.Method == "initialize" {
		res := InitializeResult("test", "0.1", httpTransport, msg.Params)
		resp, _, _ := BuildResponse(msg, res, nil)
		return resp, false, nil
	}
	return nil, true, nil
}

func TestHTTP_initialize_json(t *testing.T) {
	h := HTTPHandler(stubProc{}, HTTPConfig{Path: "/mcp", Service: "test"})
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{}}}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
	var msg Message
	if err := json.Unmarshal(rr.Body.Bytes(), &msg); err != nil {
		t.Fatal(err)
	}
	pv, _ := msg.Result.(map[string]any)["protocolVersion"].(string)
	if pv != ProtocolVersionHTTP {
		t.Fatalf("protocol %q", pv)
	}
}
