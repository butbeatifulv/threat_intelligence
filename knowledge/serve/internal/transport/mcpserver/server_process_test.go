package mcpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/butbeautifulv/veil/knowledge/serve/internal/usecase"
)

func TestProcessMessage_notification(t *testing.T) {
	srv := NewServer(usecase.NewReadUsecase(&mockExec{}), nil, nil, nil, nil, slog.Default())
	msg := rpcMessage{JSONRPC: "2.0", Method: "notifications/initialized"}
	resp, isNotification, err := srv.ProcessMessage(context.Background(), msg, false)
	if err != nil {
		t.Fatal(err)
	}
	if !isNotification || resp != nil {
		t.Fatalf("notification: resp=%v isNotification=%v", resp, isNotification)
	}
}

func TestProcessMessage_initialize(t *testing.T) {
	srv := NewServer(usecase.NewReadUsecase(&mockExec{}), nil, nil, nil, nil, slog.Default())
	msg := rpcMessage{JSONRPC: "2.0", ID: 1, Method: "initialize", Params: json.RawMessage(`{}`)}
	resp, isNotification, err := srv.ProcessMessage(context.Background(), msg, true)
	if err != nil || isNotification || resp == nil {
		t.Fatalf("err=%v notif=%v resp=%v", err, isNotification, resp)
	}
	pv, _ := resp.Result.(map[string]any)["protocolVersion"].(string)
	if pv != protocolVersionHTTP {
		t.Fatalf("protocol %q", pv)
	}
}
