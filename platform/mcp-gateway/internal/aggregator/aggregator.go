package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/butbeautifulv/veil/platform/mcp-gateway/internal/backend"
	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/mcp"
)

const (
	ServerName     = "veil-unified-mcp"
	Version        = "0.1.0"
	GraphPrefix    = "graph_"
	EngagePrefix   = "engage_"
	metaBackendKey = "x-veil-backend"
)

// Aggregator merges graph and engage Streamable HTTP MCP backends.
type Aggregator struct {
	graph  *backend.Client
	engage *backend.Client
	auth   *auth.Stack
	logger *slog.Logger
}

func New(graph, engage *backend.Client, stack *auth.Stack, logger *slog.Logger) *Aggregator {
	if logger == nil {
		logger = slog.Default()
	}
	return &Aggregator{graph: graph, engage: engage, auth: stack, logger: logger}
}

// Logger returns the aggregator logger (for HTTP wiring).
func (a *Aggregator) Logger() *slog.Logger {
	if a == nil || a.logger == nil {
		return slog.Default()
	}
	return a.logger
}

func (a *Aggregator) ProcessMessage(ctx context.Context, msg mcp.Message, httpTransport bool) (*mcp.Message, bool, error) {
	result, rerr := a.handle(ctx, msg.Method, msg.Params, httpTransport)
	return mcp.BuildResponse(msg, result, rerr)
}

func (a *Aggregator) handle(ctx context.Context, method string, params json.RawMessage, httpTransport bool) (any, error) {
	switch method {
	case "initialize":
		return mcp.InitializeResult(ServerName, Version, httpTransport, params), nil
	case "ping":
		return map[string]any{}, nil
	case "tools/list":
		return a.listTools(ctx)
	case "tools/call":
		p, err := mcp.ParseToolCallParams(params)
		if err != nil {
			return nil, err
		}
		client, bare, perm, authorize, err := a.routeTool(p.Name)
		if err != nil {
			return nil, err
		}
		ctx, err = mcp.AuthorizeToolCall(ctx, a.auth, perm, authorize)
		if err != nil {
			return nil, err
		}
		callParams, err := json.Marshal(map[string]any{
			"name":      bare,
			"arguments": p.Arguments,
		})
		if err != nil {
			return nil, err
		}
		raw, err := client.Call(ctx, "tools/call", callParams, authorizationHeader(ctx))
		if err != nil {
			if strings.HasPrefix(p.Name, EngagePrefix) {
				return nil, mcp.Errf(mcp.CodeInternal, "engage backend not configured — see UNIFIED_MCP_ENGAGE_URL: %v", err)
			}
			return nil, err
		}
		var out map[string]any
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	case "notifications/initialized":
		return nil, nil
	default:
		return nil, mcp.Errf(mcp.CodeMethodNotFound, "unknown method: %s", method)
	}
}

func (a *Aggregator) listTools(ctx context.Context) (map[string]any, error) {
	authz := authorizationHeader(ctx)
	type listed struct {
		tools []any
		err   error
	}
	var graphListed, engageListed listed
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		raw, err := a.graph.Call(ctx, "tools/list", nil, authz)
		if err != nil {
			graphListed.err = err
			return
		}
		graphListed.tools, graphListed.err = toolsFromResult(raw)
	}()
	go func() {
		defer wg.Done()
		raw, err := a.engage.Call(ctx, "tools/list", nil, authz)
		if err != nil {
			engageListed.err = err
			return
		}
		engageListed.tools, engageListed.err = toolsFromResult(raw)
	}()
	wg.Wait()
	if graphListed.err != nil && engageListed.err != nil {
		return nil, fmt.Errorf("graph tools/list: %v; engage tools/list: %v", graphListed.err, engageListed.err)
	}
	if graphListed.err != nil {
		return nil, fmt.Errorf("graph tools/list: %w", graphListed.err)
	}
	if engageListed.err != nil {
		a.logger.Warn("engage tools/list unavailable; returning graph tools only",
			slog.String("err", engageListed.err.Error()))
	}
	merged := make([]any, 0, len(graphListed.tools)+len(engageListed.tools))
	merged = append(merged, prefixTools(graphListed.tools, GraphPrefix, "graph")...)
	if engageListed.err == nil {
		merged = append(merged, prefixTools(engageListed.tools, EngagePrefix, "engage")...)
	}
	return map[string]any{"tools": merged}, nil
}

func toolsFromResult(raw json.RawMessage) ([]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	tools, ok := payload["tools"].([]any)
	if !ok {
		return nil, fmt.Errorf("missing tools array")
	}
	return tools, nil
}

func prefixTools(tools []any, prefix, backend string) []any {
	out := make([]any, 0, len(tools))
	for _, item := range tools {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		cloned := cloneMap(m)
		if name, ok := cloned["name"].(string); ok && name != "" {
			cloned["name"] = prefix + name
		}
		ann := annotationMap(cloned)
		ann[metaBackendKey] = backend
		cloned["annotations"] = ann
		out = append(out, cloned)
	}
	return out
}

func annotationMap(m map[string]any) map[string]any {
	if ann, ok := m["annotations"].(map[string]any); ok && ann != nil {
		return ann
	}
	ann := map[string]any{}
	return ann
}

func cloneMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func (a *Aggregator) routeTool(name string) (*backend.Client, string, string, func(context.Context, *auth.Stack, string) (context.Context, error), error) {
	switch {
	case strings.HasPrefix(name, GraphPrefix):
		return a.graph, strings.TrimPrefix(name, GraphPrefix), auth.PermGraphRead, auth.AuthorizeMCP, nil
	case strings.HasPrefix(name, EngagePrefix):
		return a.engage, strings.TrimPrefix(name, EngagePrefix), auth.PermEngageToolRun, auth.AuthorizeEngageMCP, nil
	default:
		return nil, "", "", nil, mcp.Errf(mcp.CodeInvalidParams, "tool name must be prefixed with %s or %s", GraphPrefix, EngagePrefix)
	}
}
