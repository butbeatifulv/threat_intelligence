package mcp

import (
	"encoding/json"
	"testing"
)

func TestNegotiateProtocol(t *testing.T) {
	if got := NegotiateProtocol(nil); got != DefaultProtocol {
		t.Fatalf("empty: %q", got)
	}
	params, _ := json.Marshal(initializeParams{ProtocolVersion: Protocol20250326})
	if got := NegotiateProtocol(params); got != Protocol20250326 {
		t.Fatalf("2025: %q", got)
	}
}

func TestInitializeResult_httpDefault(t *testing.T) {
	res := InitializeResult("svc", "1.0", true, nil)
	pv := res["protocolVersion"].(string)
	if pv != ProtocolVersionHTTP {
		t.Fatalf("got %q", pv)
	}
}

func TestNegotiateProtocol_invalidJSON(t *testing.T) {
	if got := NegotiateProtocol([]byte("{")); got != DefaultProtocol {
		t.Fatalf("got %q", got)
	}
}

func TestNegotiateProtocol_unsupportedVersion(t *testing.T) {
	params, _ := json.Marshal(initializeParams{ProtocolVersion: "2099-01-01"})
	if got := NegotiateProtocol(params); got != DefaultProtocol {
		t.Fatalf("got %q", got)
	}
}

func TestInitializeResult_stdioKeepsDefault(t *testing.T) {
	res := InitializeResult("svc", "1.0", false, nil)
	if res["protocolVersion"] != DefaultProtocol {
		t.Fatalf("got %v", res["protocolVersion"])
	}
}
