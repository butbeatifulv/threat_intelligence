package neo4jstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"

	graphneo4j "github.com/butbeautifulv/veil/knowledge/connector/neo4j"
)

type Store struct {
	client *graphneo4j.Client
}

type Config = graphneo4j.Config

func New(ctx context.Context, cfg Config) (*Store, error) {
	c, err := graphneo4j.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Store{client: c}, nil
}

func (s *Store) Close(ctx context.Context) error { return s.client.Close(ctx) }

func (s *Store) EnsureSchema(ctx context.Context) error {
	return graphneo4j.EnsureConstraints(ctx, s.client, []string{
		`CREATE CONSTRAINT package_id IF NOT EXISTS FOR (n:Package) REQUIRE n.id IS UNIQUE`,
		`CREATE CONSTRAINT security_advisory_id IF NOT EXISTS FOR (n:SecurityAdvisory) REQUIRE n.id IS UNIQUE`,
	})
}

func packageKey(ecosystem, name string) string {
	e := strings.ToLower(strings.TrimSpace(ecosystem))
	n := strings.TrimSpace(name)
	return e + ":" + n
}

func strOr(v any) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

// UpsertFromOSVVuln links existing CVE nodes to Package nodes from OSV "affected" array.
func (s *Store) UpsertFromOSVVuln(ctx context.Context, osvID string, cve string, affected []map[string]any) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	cve = strings.TrimSpace(strings.ToUpper(cve))
	if cve == "" && strings.HasPrefix(strings.ToUpper(osvID), "CVE-") {
		cve = strings.ToUpper(osvID)
	}
	for _, aff := range affected {
		pkg, _ := aff["package"].(map[string]any)
		if pkg == nil {
			continue
		}
		eco, _ := pkg["ecosystem"].(string)
		name, _ := pkg["name"].(string)
		if eco == "" || name == "" {
			continue
		}
		pid := packageKey(eco, name)
		rangesJSON, _ := json.Marshal(aff["ranges"])
		params := map[string]any{
			"pkgId":     pid,
			"ecosystem": eco,
			"name":      name,
			"purl":      strOr(pkg["purl"]),
			"ranges":    string(rangesJSON),
			"osvId":     osvID,
			"updatedAt": now,
			"cve":       cve,
		}
		q := `
MERGE (p:Package {id: $pkgId})
SET p.ecosystem = $ecosystem,
    p.name = $name,
    p.purl = CASE WHEN $purl = "" THEN p.purl ELSE $purl END,
    p.updatedAt = $updatedAt
WITH p, $cve AS cve
WHERE cve IS NOT NULL AND cve <> ""
MATCH (v:Vulnerability {cve: cve})
MERGE (v)-[r:AFFECTS_PACKAGE]->(p)
SET r.osvId = $osvId,
    r.rangesJson = $ranges,
    r.updatedAt = $updatedAt
`
		if err := s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
			_, err := tx.Run(ctx, q, params)
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}

// UpsertGHSA ingests one GitHub-reviewed advisory JSON (OSV schema 1.x).
func (s *Store) UpsertGHSA(ctx context.Context, doc map[string]any) error {
	id, _ := doc["id"].(string)
	if id == "" {
		return fmt.Errorf("ghsa: missing id")
	}
	summary, _ := doc["summary"].(string)
	modified, _ := doc["modified"].(string)
	var cweIDs []string
	if ds, ok := doc["database_specific"].(map[string]any); ok {
		if raw, ok := ds["cwe_ids"].([]any); ok {
			for _, x := range raw {
				if s, ok := x.(string); ok && s != "" {
					c := strings.TrimSpace(s)
					if !strings.HasPrefix(strings.ToUpper(c), "CWE-") {
						c = "CWE-" + strings.TrimPrefix(strings.TrimPrefix(c, "cwe-"), "CWE-")
					}
					cweIDs = append(cweIDs, c)
				}
			}
		}
	}
	var aliases []string
	if a, ok := doc["aliases"].([]any); ok {
		for _, x := range a {
			if s, ok := x.(string); ok && strings.HasPrefix(strings.ToUpper(s), "CVE-") {
				aliases = append(aliases, strings.ToUpper(s))
			}
		}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)

	params := map[string]any{
		"id":        id,
		"summary":   summary,
		"modified":  modified,
		"cves":      aliases,
		"updatedAt": now,
	}
	q := `
MERGE (a:SecurityAdvisory {id: $id})
SET a.summary = $summary,
    a.modified = CASE WHEN $modified = "" THEN a.modified ELSE $modified END,
    a.source = "github-advisory-database",
    a.updatedAt = $updatedAt
WITH a, $cves AS cves
UNWIND cves AS cve
OPTIONAL MATCH (v:Vulnerability {cve: cve})
FOREACH (_ IN CASE WHEN v IS NULL THEN [] ELSE [1] END |
  MERGE (v)-[r:HAS_ADVISORY]->(a)
  SET r.updatedAt = $updatedAt
)
`
	if err := s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	}); err != nil {
		return err
	}

	affected, _ := doc["affected"].([]any)
	for _, row := range affected {
		am, ok := row.(map[string]any)
		if !ok {
			continue
		}
		pkg, _ := am["package"].(map[string]any)
		if pkg == nil {
			continue
		}
		eco, _ := pkg["ecosystem"].(string)
		name, _ := pkg["name"].(string)
		if eco == "" || name == "" {
			continue
		}
		pid := packageKey(eco, name)
		rangesJSON, _ := json.Marshal(am["ranges"])
		p := map[string]any{
			"advId":     id,
			"pkgId":     pid,
			"eco":       eco,
			"name":      name,
			"ranges":    string(rangesJSON),
			"updatedAt": now,
		}
		qq := `
MATCH (a:SecurityAdvisory {id: $advId})
MERGE (pkg:Package {id: $pkgId})
SET pkg.ecosystem = $eco,
    pkg.name = $name,
    pkg.updatedAt = $updatedAt
MERGE (a)-[r:AFFECTS_PACKAGE]->(pkg)
SET r.rangesJson = $ranges,
    r.updatedAt = $updatedAt
`
		if err := s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
			_, err := tx.Run(ctx, qq, p)
			return err
		}); err != nil {
			return err
		}
	}

	for _, cid := range cweIDs {
		p := map[string]any{"advId": id, "cweId": cid, "updatedAt": now}
		qq := `
MATCH (a:SecurityAdvisory {id: $advId})
MERGE (c:CWE {id: $cweId})
MERGE (a)-[r:ADVISORY_MAPS_TO_CWE]->(c)
SET r.updatedAt = $updatedAt
`
		if err := s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
			_, err := tx.Run(ctx, qq, p)
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}

// ListCVEs returns up to limit CVE ids present in the graph.
func (s *Store) ListCVEs(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 100
	}
	q := `MATCH (v:Vulnerability) WHERE v.cve IS NOT NULL RETURN v.cve AS cve ORDER BY v.cve LIMIT $lim`
	raw, err := s.client.ExecRead(ctx, func(tx driver.ManagedTransaction) (any, error) {
		res, err := tx.Run(ctx, q, map[string]any{"lim": limit})
		if err != nil {
			return nil, err
		}
		var out []string
		for res.Next(ctx) {
			r := res.Record()
			if c, ok := r.Get("cve"); ok {
				if s, ok := c.(string); ok && s != "" {
					out = append(out, s)
				}
			}
		}
		return out, res.Err()
	})
	if err != nil {
		return nil, err
	}
	out, _ := raw.([]string)
	return out, nil
}
