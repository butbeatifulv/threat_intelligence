package mcpserver

import (
	"context"
	"log/slog"
	"runtime"

	"github.com/butbeautifulv/veil/knowledge/serve/internal/version"
	"github.com/butbeautifulv/veil/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
)

func (s *Server) callTool(ctx context.Context, name string, args map[string]any) (any, error) {
	ctx, span := observability.StartSpan(ctx, "mcp.tool.invoke", attribute.String("mcp.tool.name", name))
	defer span.End()
	for _, e := range allToolEntries() {
		if e.name == name {
			if e.deprecated {
				s.logger.Warn("deprecated mcp tool", slog.String("name", name))
			}
			return s.dispatchTool(ctx, name, args)
		}
	}
	s.logger.Warn("unknown tool", slog.String("name", name))
	return nil, rpcErrf(codeMethodNotFound, "unknown tool: %s", name)
}

func (s *Server) dispatchTool(ctx context.Context, name string, args map[string]any) (any, error) {
	var (
		result any
		err    error
	)
	switch name {
	case "ti_health":
		result, err = s.handleHealth(ctx)
	case "ti_list_categories":
		result, err = toolTextResult(map[string]any{"categories": s.uc.ListCategoryMeta()})
	case "ti_list_kinds_in_category":
		result, err = s.handleListKindsInCategory(ctx, args)
	case "ti_nodes_by_category":
		result, err = s.handleNodesByCategory(ctx, args)
	case "ti_search_in_category":
		result, err = s.handleSearchInCategory(ctx, args)
	case "ti_get_node":
		result, err = s.handleGetNode(ctx, args)
	case "ti_neighbors":
		result, err = s.handleNeighbors(ctx, args)
	case "ti_list_kinds":
		result, err = s.handleListKinds(ctx)
	case "ti_get_nodes_by_kind":
		result, err = s.handleNodesByKind(ctx, args)
	case "ti_search":
		result, err = s.handleSearch(ctx, args)
	case "playbook_search":
		result, err = s.handlePlaybookSearch(ctx, args)
	case "playbook_get":
		result, err = s.handlePlaybookGet(ctx, args)
	case "playbook_for_technique":
		result, err = s.handlePlaybookForTechnique(ctx, args)
	case "playbook_framework":
		result, err = s.handlePlaybookFramework(ctx, args)
	case "playbook_subdomains":
		result, err = s.handlePlaybookSubdomains(ctx, args)
	case "playbook_procedure":
		result, err = s.handlePlaybookProcedure(ctx, args)
	case "playbook_recommend_tools":
		result, err = s.handlePlaybookRecommendTools(ctx, args)
	case "playbook_ontology_subdomains":
		result, err = s.handlePlaybookOntologySubdomains(ctx, args)
	default:
		return nil, rpcErrf(codeMethodNotFound, "unknown tool: %s", name)
	}
	if err != nil {
		return nil, rpcErrf(codeToolError, "%s", err.Error())
	}
	return result, nil
}

func (s *Server) handleHealth(ctx context.Context) (any, error) {
	neo4jOK := true
	var neo4jErr string
	if err := s.uc.Ping(ctx); err != nil {
		neo4jOK = false
		neo4jErr = err.Error()
	}
	return toolTextResult(map[string]any{
		"ok":        neo4jOK,
		"service":   version.ServerName,
		"version":   version.MCP(),
		"neo4j_ok":  neo4jOK,
		"neo4j_err": neo4jErr,
		"go":        runtime.Version(),
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
	})
}

func (s *Server) handleListKindsInCategory(ctx context.Context, args map[string]any) (any, error) {
	cat := getString(args, "category")
	kinds, err := s.uc.ListKindsInCategory(ctx, cat)
	if err != nil {
		return nil, err
	}
	return toolTextResult(map[string]any{"category": cat, "kinds": kinds})
}

func (s *Server) handleNodesByCategory(ctx context.Context, args map[string]any) (any, error) {
	cat := getString(args, "category")
	kind := getString(args, "kind")
	limit := getInt(args, "limit", 200)
	offset := getInt(args, "offset", 0)
	nodes, err := s.uc.NodesByCategory(ctx, cat, kind, limit, offset)
	if err != nil {
		return nil, err
	}
	return toolTextResult(map[string]any{"category": cat, "kind": kind, "nodes": nodes})
}

func (s *Server) handleSearchInCategory(ctx context.Context, args map[string]any) (any, error) {
	cat := getString(args, "category")
	q := getString(args, "query")
	kind := getString(args, "kind")
	limit := getInt(args, "limit", 50)
	nodes, err := s.uc.SearchInCategory(ctx, cat, q, kind, limit)
	if err != nil {
		return nil, err
	}
	return toolTextResult(map[string]any{"category": cat, "query": q, "kind": kind, "nodes": nodes})
}

func (s *Server) handleGetNode(ctx context.Context, args map[string]any) (any, error) {
	id := getString(args, "id")
	n, err := s.uc.GetNode(ctx, id)
	if err != nil {
		return nil, err
	}
	return toolTextResult(map[string]any{"node": n})
}

func (s *Server) handleNeighbors(ctx context.Context, args map[string]any) (any, error) {
	id := getString(args, "id")
	depth := getInt(args, "depth", 1)
	limit := getInt(args, "limit", 500)
	g, err := s.uc.Neighbors(ctx, id, depth, limit)
	if err != nil {
		return nil, err
	}
	return toolTextResult(map[string]any{"graph": g})
}

func (s *Server) handleListKinds(ctx context.Context) (any, error) {
	kinds, err := s.uc.ListKinds(ctx)
	if err != nil {
		return nil, err
	}
	return toolTextResult(map[string]any{"kinds": kinds})
}

func (s *Server) handleNodesByKind(ctx context.Context, args map[string]any) (any, error) {
	kind := getString(args, "kind")
	limit := getInt(args, "limit", 200)
	offset := getInt(args, "offset", 0)
	nodes, err := s.uc.NodesByKind(ctx, kind, limit, offset)
	if err != nil {
		return nil, err
	}
	return toolTextResult(map[string]any{"nodes": nodes})
}

func (s *Server) handleSearch(ctx context.Context, args map[string]any) (any, error) {
	q := getString(args, "query")
	kind := getString(args, "kind")
	limit := getInt(args, "limit", 50)
	nodes, err := s.uc.Search(ctx, q, kind, limit)
	if err != nil {
		return nil, err
	}
	return toolTextResult(map[string]any{"nodes": nodes})
}
