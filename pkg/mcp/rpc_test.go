package mcp

import (
	"errors"
	"testing"

	"github.com/butbeautifulv/veil/pkg/auth"
)

func TestBuildResponse_notificationNoMethod(t *testing.T) {
	msg := Message{JSONRPC: "2.0", ID: 1}
	resp, isNotification, err := BuildResponse(msg, map[string]any{"ok": true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !isNotification || resp != nil {
		t.Fatalf("got resp=%v isNotification=%v", resp, isNotification)
	}
}

func TestBuildResponse_notificationNilID(t *testing.T) {
	msg := Message{JSONRPC: "2.0", Method: "notifications/initialized"}
	resp, isNotification, err := BuildResponse(msg, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !isNotification || resp != nil {
		t.Fatalf("got resp=%v isNotification=%v", resp, isNotification)
	}
}

func TestBuildResponse_success(t *testing.T) {
	msg := Message{JSONRPC: "2.0", ID: 42, Method: "tools/list"}
	result := map[string]any{"tools": []any{}}
	resp, isNotification, err := BuildResponse(msg, result, nil)
	if err != nil {
		t.Fatal(err)
	}
	if isNotification || resp == nil {
		t.Fatalf("resp=%v isNotification=%v", resp, isNotification)
	}
	if resp.JSONRPC != "2.0" || resp.ID != 42 {
		t.Fatalf("id/jsonrpc: %+v", resp)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	m, ok := resp.Result.(map[string]any)
	if !ok || m["tools"] == nil {
		t.Fatalf("result: %+v", resp.Result)
	}
}

func TestBuildResponse_error(t *testing.T) {
	msg := Message{JSONRPC: "2.0", ID: "req-1", Method: "tools/call"}
	rerr := Err(CodeInvalidParams, "bad tool")
	resp, isNotification, err := BuildResponse(msg, nil, rerr)
	if err != nil {
		t.Fatal(err)
	}
	if isNotification || resp == nil {
		t.Fatalf("resp=%v isNotification=%v", resp, isNotification)
	}
	if resp.Result != nil {
		t.Fatalf("expected no result: %+v", resp)
	}
	if resp.Error == nil || resp.Error.Code != CodeInvalidParams {
		t.Fatalf("error: %+v", resp.Error)
	}
}

func TestBuildResponse_mapsAuthErrors(t *testing.T) {
	msg := Message{JSONRPC: "2.0", ID: 1, Method: "tools/call"}
	resp, _, err := BuildResponse(msg, nil, auth.ErrUnauthorized)
	if err != nil {
		t.Fatal(err)
	}
	var re *RPCError
	if !errors.As(resp.Error, &re) || re.Code != CodeAuthError {
		t.Fatalf("error: %+v", resp.Error)
	}
}
