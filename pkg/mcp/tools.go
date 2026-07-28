package mcp

import "encoding/json"

// ToolDescriptor is a minimal tools/list entry.
type ToolDescriptor struct {
	Name        string
	Description string
	InputSchema map[string]any
}

// ListToolsPayload builds the tools/list result object.
func ListToolsPayload(tools []ToolDescriptor) map[string]any {
	out := make([]map[string]any, 0, len(tools))
	for _, t := range tools {
		out = append(out, map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"inputSchema": t.InputSchema,
		})
	}
	return map[string]any{"tools": out}
}

// ToolCallParams holds tools/call arguments.
type ToolCallParams struct {
	Name      string
	Arguments map[string]any
}

// ParseToolCallParams decodes tools/call params.
func ParseToolCallParams(params json.RawMessage) (ToolCallParams, error) {
	var p struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return ToolCallParams{}, Errf(CodeInvalidParams, "bad params: %v", err)
	}
	if p.Name == "" {
		return ToolCallParams{}, Err(CodeInvalidParams, "tool name required")
	}
	return ToolCallParams{Name: p.Name, Arguments: p.Arguments}, nil
}
