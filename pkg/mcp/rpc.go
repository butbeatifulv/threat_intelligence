package mcp

import "context"

// Processor handles one inbound MCP JSON-RPC message.
type Processor interface {
	ProcessMessage(ctx context.Context, msg Message, httpTransport bool) (resp *Message, isNotification bool, err error)
}

// BuildResponse wraps a handler result as a JSON-RPC response (or notification skip).
func BuildResponse(msg Message, result any, rerr error) (resp *Message, isNotification bool, err error) {
	if msg.Method == "" {
		return nil, true, nil
	}
	if msg.ID == nil {
		return nil, true, nil
	}
	out := &Message{JSONRPC: "2.0", ID: msg.ID}
	if rerr != nil {
		out.Error = ToRPCError(rerr)
	} else {
		out.Result = result
	}
	return out, false, nil
}
