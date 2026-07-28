package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"
)

type mockProcessor struct {
	err error
}

func (m mockProcessor) ProcessMessage(ctx context.Context, msg Message, httpTransport bool) (*Message, bool, error) {
	if m.err != nil {
		return nil, false, m.err
	}
	if msg.Method == "" {
		return nil, true, nil
	}
	return &Message{JSONRPC: "2.0", ID: msg.ID, Result: map[string]any{"ok": true}}, false, nil
}

func writeFramed(w *bytes.Buffer, payload []byte) {
	fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(payload))
	w.Write(payload)
}

func TestRunStdio_invalidReaders(t *testing.T) {
	if err := RunStdio(context.Background(), mockProcessor{}, 1, 2); err == nil {
		t.Fatal("expected invalid reader error")
	}
	if err := RunStdio(context.Background(), mockProcessor{}, &bytes.Buffer{}, 1); err == nil {
		t.Fatal("expected invalid writer error")
	}
}

func TestRunStdio_parseErrorAndRequest(t *testing.T) {
	var in bytes.Buffer
	writeFramed(&in, []byte("{invalid"))
	req := Message{JSONRPC: "2.0", ID: 1, Method: "tools/list"}
	b, _ := json.Marshal(req)
	writeFramed(&in, b)

	var out bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- RunStdio(ctx, mockProcessor{}, &in, &out) }()
	<-done
	if out.Len() == 0 {
		t.Fatal("expected response bytes")
	}
}

func TestRunStdio_ctxCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := RunStdio(ctx, mockProcessor{}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunStdio_processorError(t *testing.T) {
	var in bytes.Buffer
	req := Message{JSONRPC: "2.0", ID: 1, Method: "fail"}
	b, _ := json.Marshal(req)
	writeFramed(&in, b)
	err := RunStdio(context.Background(), mockProcessor{err: io.EOF}, &in, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected processor error")
	}
}

func TestRunStdio_eof(t *testing.T) {
	err := RunStdio(context.Background(), mockProcessor{}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
}
