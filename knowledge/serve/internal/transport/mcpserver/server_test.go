package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/butbeautifulv/veil/knowledge/serve/internal/usecase"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/version"
)

func TestServer_initialize_tools_ping(t *testing.T) {
	uc := usecase.NewReadUsecase(&mockExec{})
	srv := NewServer(uc, nil, nil, nil, nil, slog.Default())

	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Run(ctx, stdinR, stdoutW)
	}()

	writeReq := func(id int, method string, params any) {
		t.Helper()
		var rawParams json.RawMessage
		if params != nil {
			b, _ := json.Marshal(params)
			rawParams = b
		}
		rw := newFramedRW(strings.NewReader(""), stdinW)
		if err := rw.WriteJSON(ctx, rpcMessage{
			JSONRPC: "2.0",
			ID:      id,
			Method:  method,
			Params:  rawParams,
		}); err != nil {
			t.Fatal(err)
		}
	}

	readResp := func() rpcMessage {
		t.Helper()
		rw := newFramedRW(stdoutR, io.Discard)
		payload, err := rw.Read(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var msg rpcMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			t.Fatal(err)
		}
		return msg
	}

	writeReq(1, "initialize", map[string]any{
		"protocolVersion": protocol20241105,
		"capabilities":    map[string]any{},
	})
	init := readResp()
	if init.Error != nil {
		t.Fatalf("initialize error: %+v", init.Error)
	}
	info, _ := init.Result.(map[string]any)["serverInfo"].(map[string]any)
	if info["name"] != version.ServerName {
		t.Fatalf("server name: %v", info["name"])
	}
	pv, _ := init.Result.(map[string]any)["protocolVersion"].(string)
	if pv != protocol20241105 {
		t.Fatalf("protocol version: %v", pv)
	}

	writeReq(2, "ping", map[string]any{})
	ping := readResp()
	if ping.Error != nil {
		t.Fatalf("ping error: %+v", ping.Error)
	}

	writeReq(3, "tools/list", nil)
	tools := readResp()
	if tools.Error != nil {
		t.Fatalf("tools/list error: %+v", tools.Error)
	}
	result, _ := tools.Result.(map[string]any)
	list, _ := result["tools"].([]any)
	if len(list) < 7 {
		t.Fatalf("expected categorical + legacy tools, got %d", len(list))
	}

	writeReq(4, "tools/call", map[string]any{
		"name":      "ti_list_categories",
		"arguments": map[string]any{},
	})
	call := readResp()
	if call.Error != nil {
		t.Fatalf("tools/call error: %+v", call.Error)
	}

	writeReq(5, "tools/call", map[string]any{
		"name":      "ti_health",
		"arguments": map[string]any{},
	})
	health := readResp()
	if health.Error != nil {
		t.Fatalf("health error: %+v", health.Error)
	}

	cancel()
	_ = stdinW.Close()
	_ = stdoutW.Close()
	select {
	case err := <-errCh:
		if err != nil && err != io.EOF && !isClosedErr(err) {
			t.Logf("run exit: %v", err)
		}
	case <-time.After(2 * time.Second):
	}
}

func TestServer_unknownMethod(t *testing.T) {
	uc := usecase.NewReadUsecase(&mockExec{})
	srv := NewServer(uc, nil, nil, nil, nil, slog.Default())

	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()
	ctx := context.Background()
	go func() { _ = srv.Run(ctx, stdinR, stdoutW) }()

	rw := newFramedRW(strings.NewReader(""), stdinW)
	_ = rw.WriteJSON(ctx, rpcMessage{
		JSONRPC: "2.0",
		ID:      9,
		Method:  "nope/method",
	})
	rwOut := newFramedRW(stdoutR, io.Discard)
	payload, err := rwOut.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var msg rpcMessage
	_ = json.Unmarshal(payload, &msg)
	if msg.Error == nil || msg.Error.Code != codeMethodNotFound {
		t.Fatalf("expected method not found, got %+v", msg.Error)
	}
	_ = stdinW.Close()
	_ = stdoutW.Close()
}

func isClosedErr(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "io: read/write on closed pipe" ||
		bytes.Contains([]byte(err.Error()), []byte("closed"))
}
