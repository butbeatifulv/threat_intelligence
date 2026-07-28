package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/butbeautifulv/veil/pkg/mcp"
)

// Client forwards JSON-RPC to a Streamable HTTP MCP backend.
type Client struct {
	Name string
	URL  string
	HTTP *http.Client
}

// Call performs one MCP request and returns the result payload.
func (c *Client) Call(ctx context.Context, method string, params json.RawMessage, authorization string) (json.RawMessage, error) {
	if c == nil || c.URL == "" {
		name := "unknown"
		if c != nil {
			name = c.Name
		}
		return nil, fmt.Errorf("backend %q: URL not configured", name)
	}
	reqBody, err := json.Marshal(mcp.Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	})
	if err != nil {
		return nil, err
	}
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("backend %q: HTTP %d: %s", c.Name, resp.StatusCode, trimBody(body))
	}
	var msg mcp.Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("backend %q: decode response: %w", c.Name, err)
	}
	if msg.Error != nil {
		return nil, &mcp.RPCError{Code: msg.Error.Code, Message: msg.Error.Message}
	}
	raw, err := json.Marshal(msg.Result)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func trimBody(b []byte) string {
	const max = 256
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "…"
}
