package query

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// ReadExecutor runs read-only Cypher (implemented by graph/neo4j Client and mcp neo4jconn.Connector).
type ReadExecutor interface {
	ExecRead(ctx context.Context, fn func(tx driver.ManagedTransaction) (any, error)) (any, error)
}

// Service runs categorized graph reads.
type Service struct {
	exec ReadExecutor
}

func NewService(exec ReadExecutor) *Service {
	return &Service{exec: exec}
}

// Ping verifies Neo4j read connectivity.
func (s *Service) Ping(ctx context.Context) error {
	_, err := s.exec.ExecRead(ctx, func(tx driver.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, `RETURN 1`, nil)
		return nil, err
	})
	return err
}

// nodeTextSearchPredicate matches common TI fields plus engage scan metadata.
// Neo4j 5+ requires IS NOT NULL (exists(n.prop) predicate syntax removed).
const nodeTextSearchPredicate = `
  (n.title IS NOT NULL AND toLower(n.title) CONTAINS $q) OR
  (n.name IS NOT NULL AND toLower(n.name) CONTAINS $q) OR
  (n.id IS NOT NULL AND toLower(toString(n.id)) CONTAINS $q) OR
  (n.cve IS NOT NULL AND toLower(n.cve) CONTAINS $q) OR
  (n.value IS NOT NULL AND toLower(n.value) CONTAINS $q) OR
  (n.uri IS NOT NULL AND toLower(n.uri) CONTAINS $q) OR
  (n.link IS NOT NULL AND toLower(n.link) CONTAINS $q) OR
  (n.target IS NOT NULL AND toLower(n.target) CONTAINS $q) OR
  (n.tool IS NOT NULL AND toLower(n.tool) CONTAINS $q) OR
  (n.severity IS NOT NULL AND toLower(n.severity) CONTAINS $q) OR
  (n.description IS NOT NULL AND toLower(n.description) CONTAINS $q)
`

const nodeMatchByID = `elementId(n) = $id OR n.id = $id OR n.cve = $id OR n.uri = $id OR n.link = $id OR (n:EngageTarget AND n.name = $id)`

const seedMatchByID = `elementId(seed) = $id OR seed.id = $id OR seed.cve = $id OR seed.uri = $id OR seed.link = $id OR (seed:EngageTarget AND seed.name = $id)`

// ListKinds returns all distinct labels sorted (legacy / discovery).
func (s *Service) ListKinds(ctx context.Context) ([]string, error) {
	res, err := s.exec.ExecRead(ctx, func(tx driver.ManagedTransaction) (any, error) {
		r, err := tx.Run(ctx, listKindsCypher, nil)
		if err != nil {
			return nil, err
		}
		var out []string
		for r.Next(ctx) {
			out = append(out, r.Record().Values[0].(string))
		}
		return out, r.Err()
	})
	if err != nil {
		return nil, err
	}
	return res.([]string), nil
}

// ListKindsInCategory returns labels from the category that appear on at least one node, with counts.
func (s *Service) ListKindsInCategory(ctx context.Context, category string) ([]KindCount, error) {
	allowed, ok := labelsForCategory(category)
	if !ok {
		return nil, fmt.Errorf("unknown category: %s", category)
	}
	res, err := s.exec.ExecRead(ctx, func(tx driver.ManagedTransaction) (any, error) {
		r, err := tx.Run(ctx, listKindsInCategoryCypher, map[string]any{"allowed": allowed})
		if err != nil {
			return nil, err
		}
		var out []KindCount
		for r.Next(ctx) {
			rec := r.Record()
			k, _ := rec.Get("kind")
			c, _ := rec.Get("c")
			out = append(out, KindCount{Kind: toString(k), Count: toInt64(c)})
		}
		return out, r.Err()
	})
	if err != nil {
		return nil, err
	}
	return res.([]KindCount), nil
}

// NodesByKind matches a single label.
func (s *Service) NodesByKind(ctx context.Context, kind string, limit, offset int) ([]Node, error) {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return nil, fmt.Errorf("kind is empty")
	}
	limit = clampLimit(limit, 200, 5000)
	offset = clampOffset(offset)
	q := buildNodesByKindCypher(kind)
	return s.runNodesQuery(ctx, q, map[string]any{"limit": limit, "offset": offset})
}

// NodesByCategory lists nodes of kind within a category (kind must belong to category).
func (s *Service) NodesByCategory(ctx context.Context, category, kind string, limit, offset int) ([]Node, error) {
	allowed, ok := labelsForCategory(category)
	if !ok {
		return nil, fmt.Errorf("unknown category: %s", category)
	}
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return nil, fmt.Errorf("kind is empty")
	}
	if !labelInList(kind, allowed) {
		return nil, fmt.Errorf("kind %q is not in category %q", kind, category)
	}
	return s.NodesByKind(ctx, kind, limit, offset)
}

func labelInList(kind string, allowed []string) bool {
	for _, a := range allowed {
		if a == kind {
			return true
		}
	}
	return false
}

func (s *Service) runNodesQuery(ctx context.Context, q string, params map[string]any) ([]Node, error) {
	res, err := s.exec.ExecRead(ctx, func(tx driver.ManagedTransaction) (any, error) {
		r, err := tx.Run(ctx, q, params)
		if err != nil {
			return nil, err
		}
		var out []Node
		for r.Next(ctx) {
			rec := r.Record()
			id, _ := rec.Get("id")
			labels, _ := rec.Get("labels")
			props, _ := rec.Get("props")
			out = append(out, Node{
				ElementID: toString(id),
				Labels:    toStringSlice(labels),
				Props:     toMap(props),
			})
		}
		return out, r.Err()
	})
	if err != nil {
		return nil, err
	}
	return res.([]Node), nil
}

// GetNode resolves by elementId or common business keys.
func (s *Service) GetNode(ctx context.Context, id string) (*Node, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id is empty")
	}
	q := buildGetNodeCypher()
	res, err := s.exec.ExecRead(ctx, func(tx driver.ManagedTransaction) (any, error) {
		r, err := tx.Run(ctx, q, map[string]any{"id": id})
		if err != nil {
			return nil, err
		}
		if !r.Next(ctx) {
			if err := r.Err(); err != nil {
				return nil, err
			}
			return (*Node)(nil), nil
		}
		rec := r.Record()
		n := &Node{
			ElementID: toString(mustGet(rec, "id")),
			Labels:    toStringSlice(mustGet(rec, "labels")),
			Props:     toMap(mustGet(rec, "props")),
		}
		return n, r.Err()
	})
	if err != nil {
		return nil, err
	}
	return res.(*Node), nil
}

// Neighbors returns a k-hop subgraph around a node.
func (s *Service) Neighbors(ctx context.Context, id string, depth, limit int) (*Graph, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id is empty")
	}
	depth = clampDepth(depth)
	limit = clampLimit(limit, 500, 5000)
	res, err := s.exec.ExecRead(ctx, func(tx driver.ManagedTransaction) (any, error) {
		r, err := tx.Run(ctx, buildNeighborsNodesCypher(depth), map[string]any{"id": id, "limit": limit})
		if err != nil {
			return nil, err
		}

		nodeSeen := map[string]Node{}
		for r.Next(ctx) {
			rec := r.Record()
			nid := toString(mustGet(rec, "id"))
			if nid != "" {
				if _, ok := nodeSeen[nid]; !ok {
					nodeSeen[nid] = Node{
						ElementID: nid,
						Labels:    toStringSlice(mustGet(rec, "labels")),
						Props:     toMap(mustGet(rec, "props")),
					}
				}
			}
		}
		if err := r.Err(); err != nil {
			return nil, err
		}

		nodes := make([]Node, 0, len(nodeSeen))
		for _, n := range nodeSeen {
			nodes = append(nodes, n)
		}

		edgesRes, err := tx.Run(ctx, buildNeighborsEdgesCypher(depth), map[string]any{"id": id, "limit": limit})
		if err != nil {
			return nil, err
		}
		var edges []Edge
		for edgesRes.Next(ctx) {
			rec := edgesRes.Record()
			edges = append(edges, Edge{
				ElementID: toString(mustGet(rec, "id")),
				Type:      toString(mustGet(rec, "type")),
				Source:    toString(mustGet(rec, "source")),
				Target:    toString(mustGet(rec, "target")),
			})
		}
		if err := edgesRes.Err(); err != nil {
			return nil, err
		}
		return &Graph{Nodes: nodes, Edges: edges}, nil
	})
	if err != nil {
		return nil, err
	}
	return res.(*Graph), nil
}

// Search substring over common text fields; optional kind filter (single label).
func (s *Service) Search(ctx context.Context, query, kind string, limit int) ([]Node, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is empty")
	}
	limit = clampLimit(limit, 50, 2000)
	kind = strings.TrimSpace(kind)
	params := map[string]any{"q": strings.ToLower(query), "limit": limit}
	return s.runNodesQuery(ctx, buildSearchCypher(kind), params)
}

// SearchInCategory restricts search to nodes that have at least one label from the category.
func (s *Service) SearchInCategory(ctx context.Context, category, queryText, kind string, limit int) ([]Node, error) {
	allowed, ok := labelsForCategory(category)
	if !ok {
		return nil, fmt.Errorf("unknown category: %s", category)
	}
	queryText = strings.TrimSpace(queryText)
	if queryText == "" {
		return nil, fmt.Errorf("query is empty")
	}
	limit = clampLimit(limit, 50, 2000)
	kind = strings.TrimSpace(kind)
	if kind != "" && !labelInList(kind, allowed) {
		return nil, fmt.Errorf("kind %q is not in category %q", kind, category)
	}
	params := map[string]any{"q": strings.ToLower(queryText), "limit": limit, "allowed": allowed}
	return s.runNodesQuery(ctx, buildSearchInCategoryCypher(kind), params)
}

func safeLabel(label string) string {
	var b strings.Builder
	for _, r := range label {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return "Node"
	}
	return out
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toInt64(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	case float64:
		return int64(x)
	default:
		return 0
	}
}

func toStringSlice(v any) []string {
	a, ok := v.([]any)
	if !ok {
		if ss, ok := v.([]string); ok {
			return ss
		}
		return nil
	}
	out := make([]string, 0, len(a))
	for _, x := range a {
		if s, ok := x.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func toMap(v any) map[string]any {
	m, ok := v.(map[string]any)
	if ok {
		return m
	}
	b, err := json.Marshal(v)
	if err != nil {
		return map[string]any{"value": fmt.Sprintf("%v", v)}
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return map[string]any{"value": fmt.Sprintf("%v", v)}
	}
	return out
}

func mustGet(rec *driver.Record, key string) any {
	v, _ := rec.Get(key)
	return v
}
