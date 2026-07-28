package query

// Node is a Neo4j node projection for API/MCP.
type Node struct {
	ElementID string         `json:"id"`
	Labels    []string       `json:"labels"`
	Props     map[string]any `json:"props"`
}

// Edge is a Neo4j relationship projection.
type Edge struct {
	ElementID string `json:"id"`
	Type      string `json:"type"`
	Source    string `json:"source"`
	Target    string `json:"target"`
}

// Graph is a small subgraph payload.
type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// KindCount is a label with how many nodes carry it (any label match).
type KindCount struct {
	Kind  string `json:"kind"`
	Count int64  `json:"count"`
}

// CategoryMeta describes a stable product category for routing queries.
type CategoryMeta struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Labels      []string `json:"labels"`
}
