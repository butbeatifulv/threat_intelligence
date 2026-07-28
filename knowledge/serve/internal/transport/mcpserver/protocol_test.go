package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestFramedRW_roundtrip(t *testing.T) {
	var buf bytes.Buffer
	rw := newFramedRW(strings.NewReader(""), &buf)

	msg := rpcMessage{JSONRPC: "2.0", ID: 1, Result: map[string]any{"ok": true}}
	if err := rw.WriteJSON(context.Background(), msg); err != nil {
		t.Fatal(err)
	}

	rw2 := newFramedRW(bytes.NewReader(buf.Bytes()), io.Discard)
	payload, err := rw2.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var got rpcMessage
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatal(err)
	}
	if got.JSONRPC != "2.0" || got.ID != float64(1) {
		t.Fatalf("unexpected message: %+v", got)
	}
}
