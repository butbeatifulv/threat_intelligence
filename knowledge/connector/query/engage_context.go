package query

import (
	"context"

	"github.com/butbeautifulv/veil/pkg/engage/hostnorm"
	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// EngageFindingContext is a finding with linked vulnerability nodes.
type EngageFindingContext struct {
	Node                   Node   `json:"node"`
	RelatedVulnerabilities []Node `json:"related_vulnerabilities"`
}

// EngageTargetContext is the read model for an engage target host subgraph.
type EngageTargetContext struct {
	Host            string                 `json:"host"`
	Target          *Node                  `json:"target,omitempty"`
	ToolRuns        []Node                 `json:"tool_runs"`
	Findings        []EngageFindingContext `json:"findings"`
	Vulnerabilities []Node                 `json:"vulnerabilities"`
}

// EngageTargetContext loads EngageTarget, tool runs, findings, and MAY_RELATE_TO CVE links.
func (s *Service) EngageTargetContext(ctx context.Context, host string) (*EngageTargetContext, error) {
	host = hostnorm.NormalizeHost(host)
	if host == "" {
		return nil, nil
	}
	res, err := s.exec.ExecRead(ctx, func(tx driver.ManagedTransaction) (any, error) {
		out := &EngageTargetContext{Host: host}
		targetQ := `
MATCH (t:EngageTarget {name: $host})
RETURN elementId(t) AS id, labels(t) AS labels, properties(t) AS props
LIMIT 1`
		tr, err := tx.Run(ctx, targetQ, map[string]any{"host": host})
		if err != nil {
			return nil, err
		}
		if tr.Next(ctx) {
			rec := tr.Record()
			n := Node{
				ElementID: toString(mustGet(rec, "id")),
				Labels:    toStringSlice(mustGet(rec, "labels")),
				Props:     toMap(mustGet(rec, "props")),
			}
			out.Target = &n
		}
		if err := tr.Err(); err != nil {
			return nil, err
		}
		if out.Target == nil {
			return out, nil
		}

		runsQ := `
MATCH (t:EngageTarget {name: $host})-[:ENGAGE_RAN]->(r:EngageToolRun)
RETURN elementId(r) AS id, labels(r) AS labels, properties(r) AS props`
		rr, err := tx.Run(ctx, runsQ, map[string]any{"host": host})
		if err != nil {
			return nil, err
		}
		for rr.Next(ctx) {
			rec := rr.Record()
			out.ToolRuns = append(out.ToolRuns, Node{
				ElementID: toString(mustGet(rec, "id")),
				Labels:    toStringSlice(mustGet(rec, "labels")),
				Props:     toMap(mustGet(rec, "props")),
			})
		}
		if err := rr.Err(); err != nil {
			return nil, err
		}

		findingsQ := `
MATCH (t:EngageTarget {name: $host})-[:ENGAGE_FOUND]->(f:EngageFinding)
OPTIONAL MATCH (f)-[:MAY_RELATE_TO]->(v:Vulnerability)
RETURN elementId(f) AS fid, labels(f) AS flabels, properties(f) AS fprops,
       elementId(v) AS vid, labels(v) AS vlabels, properties(v) AS vprops`
		fr, err := tx.Run(ctx, findingsQ, map[string]any{"host": host})
		if err != nil {
			return nil, err
		}
		vulnSeen := map[string]Node{}
		findingByID := map[string]*EngageFindingContext{}
		var findingOrder []string
		for fr.Next(ctx) {
			rec := fr.Record()
			fid := toString(mustGet(rec, "fid"))
			if fid == "" {
				continue
			}
			fc, ok := findingByID[fid]
			if !ok {
				fc = &EngageFindingContext{
					Node: Node{
						ElementID: fid,
						Labels:    toStringSlice(mustGet(rec, "flabels")),
						Props:     toMap(mustGet(rec, "fprops")),
					},
				}
				findingByID[fid] = fc
				findingOrder = append(findingOrder, fid)
			}
			vid := toString(mustGet(rec, "vid"))
			if vid == "" {
				continue
			}
			vn := Node{
				ElementID: vid,
				Labels:    toStringSlice(mustGet(rec, "vlabels")),
				Props:     toMap(mustGet(rec, "vprops")),
			}
			fc.RelatedVulnerabilities = append(fc.RelatedVulnerabilities, vn)
			vulnSeen[vid] = vn
		}
		if err := fr.Err(); err != nil {
			return nil, err
		}
		for _, fid := range findingOrder {
			out.Findings = append(out.Findings, *findingByID[fid])
		}
		for _, v := range vulnSeen {
			out.Vulnerabilities = append(out.Vulnerabilities, v)
		}
		return out, nil
	})
	if err != nil {
		return nil, err
	}
	return res.(*EngageTargetContext), nil
}

// NormalizeEngageHost strips scheme/path for EngageTarget.name lookup.
func NormalizeEngageHost(target string) string {
	return hostnorm.NormalizeHost(target)
}
