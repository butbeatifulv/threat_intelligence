package mcpserver

import "encoding/json"

func toolTextResult(v any) (any, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"content": []any{
			map[string]any{"type": "text", "text": string(b)},
		},
	}, nil
}
