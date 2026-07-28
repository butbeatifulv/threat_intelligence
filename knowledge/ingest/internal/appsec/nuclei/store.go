package neo4jstore

import (
	"context"
	"fmt"
	"strings"
	"time"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"

	graphneo4j "github.com/butbeautifulv/veil/knowledge/connector/neo4j"
)

type Config = graphneo4j.Config

type Store struct {
	client *graphneo4j.Client
}

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
		`CREATE CONSTRAINT nuclei_template_id IF NOT EXISTS FOR (n:NucleiTemplate) REQUIRE n.id IS UNIQUE`,
	})
}

func StableID(prefix, key string) string {
	h := fmt.Sprintf("%s:%s", prefix, key)
	var x uint64 = 14695981039346656037
	for _, b := range []byte(h) {
		x ^= uint64(b)
		x *= 1099511628211
	}
	return fmt.Sprintf("%s:%016x", prefix, x)
}

func (s *Store) UpsertNucleiTemplate(ctx context.Context, id, templateKey, path, name, severity, tagsJSON, cveID, cweID, markdown string) error {
	if id == "" {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	cveID = strings.TrimSpace(strings.ToUpper(cveID))
	cweID = strings.TrimSpace(cweID)
	if cweID != "" && !strings.HasPrefix(strings.ToUpper(cweID), "CWE-") {
		cweID = "CWE-" + strings.TrimPrefix(strings.TrimPrefix(cweID, "cwe-"), "CWE-")
	}
	params := map[string]any{
		"id": id, "key": templateKey, "path": path, "name": name, "severity": severity,
		"tags": tagsJSON, "cve": cveID, "cwe": cweID, "md": markdown, "ts": now,
	}
	q := `
MERGE (n:NucleiTemplate {id: $id})
SET n.templateKey = $key,
    n.path = $path,
    n.name = $name,
    n.severity = $severity,
    n.tags = $tags,
    n.cve = CASE WHEN $cve = "" THEN n.cve ELSE $cve END,
    n.cwe = CASE WHEN $cwe = "" THEN n.cwe ELSE $cwe END,
    n.markdown = $md,
    n.source = "nuclei-templates",
    n.updatedAt = $ts
`
	if err := s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	}); err != nil {
		return err
	}
	if cveID != "" {
		qq := `
MATCH (n:NucleiTemplate {id: $id})
WITH n
OPTIONAL MATCH (v:Vulnerability {cve: $cve})
FOREACH (_ IN CASE WHEN v IS NULL THEN [] ELSE [1] END |
  MERGE (n)-[r:RELATES_TO_CVE]->(v)
  SET r.updatedAt = $ts
)`
		if err := s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
			_, err := tx.Run(ctx, qq, map[string]any{"id": id, "cve": cveID, "ts": now})
			return err
		}); err != nil {
			return err
		}
	}
	if cweID != "" {
		qq := `
MATCH (n:NucleiTemplate {id: $id})
MERGE (c:CWE {id: $cwe})
MERGE (n)-[r:MAPS_TO_CWE]->(c)
SET r.updatedAt = $ts
`
		if err := s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
			_, err := tx.Run(ctx, qq, map[string]any{"id": id, "cwe": cweID, "ts": now})
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}
