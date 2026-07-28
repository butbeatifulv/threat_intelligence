package query

import "fmt"

// listKindsCypher returns distinct node labels (discovery).
const listKindsCypher = `MATCH (n) UNWIND labels(n) AS l RETURN DISTINCT l AS label ORDER BY label`

// listKindsInCategoryCypher counts labels within a category allowlist.
const listKindsInCategoryCypher = `
MATCH (n)
UNWIND labels(n) AS l
WHERE l IN $allowed
RETURN l AS kind, count(DISTINCT n) AS c
ORDER BY c DESC
`

func buildNodesByKindCypher(kind string) string {
	return fmt.Sprintf(`
MATCH (n:%s)
WITH n SKIP $offset LIMIT $limit
RETURN elementId(n) AS id, labels(n) AS labels, properties(n) AS props
`, safeLabel(kind))
}

func buildGetNodeCypher() string {
	return `
MATCH (n)
WHERE ` + nodeMatchByID + `
RETURN elementId(n) AS id, labels(n) AS labels, properties(n) AS props
LIMIT 1`
}

func buildNeighborsNodesCypher(depth int) string {
	return fmt.Sprintf(`
MATCH (seed)
WHERE %s
WITH seed
MATCH p=(seed)-[r*1..%d]-(n)
WITH collect(DISTINCT seed) + collect(DISTINCT n) AS ns, relationships(p) AS rs
UNWIND ns AS node
WITH DISTINCT node, rs
RETURN elementId(node) AS id, labels(node) AS labels, properties(node) AS props, rs AS rels
LIMIT $limit
`, seedMatchByID, depth)
}

func buildNeighborsEdgesCypher(depth int) string {
	return fmt.Sprintf(`
MATCH (seed)
WHERE %s
MATCH p=(seed)-[r*1..%d]-(n)
UNWIND relationships(p) AS rel
RETURN DISTINCT elementId(rel) AS id, type(rel) AS type, elementId(startNode(rel)) AS source, elementId(endNode(rel)) AS target
LIMIT $limit
`, seedMatchByID, depth)
}

func buildSearchCypher(kind string) string {
	if kind != "" {
		return fmt.Sprintf(`
MATCH (n:%s)
WHERE (`+nodeTextSearchPredicate+`)
RETURN elementId(n) AS id, labels(n) AS labels, properties(n) AS props
LIMIT $limit
`, safeLabel(kind))
	}
	return `
MATCH (n)
WHERE (` + nodeTextSearchPredicate + `)
RETURN elementId(n) AS id, labels(n) AS labels, properties(n) AS props
LIMIT $limit
`
}

func buildSearchInCategoryCypher(kind string) string {
	if kind != "" {
		return fmt.Sprintf(`
MATCH (n:%s)
WHERE any(l IN labels(n) WHERE l IN $allowed)
AND (`+nodeTextSearchPredicate+`)
RETURN elementId(n) AS id, labels(n) AS labels, properties(n) AS props
LIMIT $limit
`, safeLabel(kind))
	}
	return `
MATCH (n)
WHERE any(l IN labels(n) WHERE l IN $allowed)
AND (` + nodeTextSearchPredicate + `)
RETURN elementId(n) AS id, labels(n) AS labels, properties(n) AS props
LIMIT $limit
`
}

func clampLimit(limit, defaultVal, max int) int {
	if limit <= 0 || limit > max {
		return defaultVal
	}
	return limit
}

func clampOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func clampDepth(depth int) int {
	if depth <= 0 || depth > 3 {
		return 1
	}
	return depth
}
