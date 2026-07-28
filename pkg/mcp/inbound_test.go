package mcp

import (
	"encoding/json"
	"testing"
)

func TestParseInboundMessages_singleAndBatch(t *testing.T) {
	one := Message{JSONRPC: "2.0", ID: 1, Method: "ping"}
	b, _ := json.Marshal(one)
	msgs, err := ParseInboundMessages(b)
	if err != nil || len(msgs) != 1 || msgs[0].Method != "ping" {
		t.Fatalf("single: %v %v", msgs, err)
	}
	batch, _ := json.Marshal([]Message{{Method: "a"}, {Method: "b"}})
	msgs2, err := ParseInboundMessages(batch)
	if err != nil || len(msgs2) != 2 {
		t.Fatalf("batch: %v", msgs2)
	}
	_, err = ParseInboundMessages([]byte("not json"))
	if err == nil {
		t.Fatal("expected parse error")
	}
	_, err = ParseInboundMessages([]byte("[not json"))
	if err == nil {
		t.Fatal("expected batch parse error")
	}
}
