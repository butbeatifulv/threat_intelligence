package mcp

import "encoding/json"

const (
	Protocol20241105    = "2024-11-05"
	Protocol20250326    = "2025-03-26"
	DefaultProtocol     = Protocol20241105
	ProtocolVersionHTTP = Protocol20250326
)

var supportedProtocols = map[string]struct{}{
	Protocol20241105: {},
	Protocol20250326: {},
}

type initializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ClientInfo      map[string]any `json:"clientInfo"`
}

// NegotiateProtocol picks a supported MCP protocol version from initialize params.
func NegotiateProtocol(params json.RawMessage) string {
	if len(params) == 0 {
		return DefaultProtocol
	}
	var p initializeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return DefaultProtocol
	}
	if p.ProtocolVersion == "" {
		return DefaultProtocol
	}
	if _, ok := supportedProtocols[p.ProtocolVersion]; ok {
		return p.ProtocolVersion
	}
	return DefaultProtocol
}

// InitializeResult builds the initialize response payload.
func InitializeResult(serverName, version string, httpTransport bool, params json.RawMessage) map[string]any {
	pv := NegotiateProtocol(params)
	if httpTransport && pv == DefaultProtocol {
		pv = ProtocolVersionHTTP
	}
	return map[string]any{
		"protocolVersion": pv,
		"serverInfo": map[string]any{
			"name":    serverName,
			"version": version,
		},
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
	}
}
