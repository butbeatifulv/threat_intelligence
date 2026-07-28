package neo4jstore

import (
	"context"
	"strings"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/butbeautifulv/veil/knowledge/connector/neo4j"
)

type Config = neo4j.Config

type Store struct {
	client *neo4j.Client
}

func New(ctx context.Context, cfg Config) (*Store, error) {
	c, err := neo4j.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Store{client: c}, nil
}

func (s *Store) Close(ctx context.Context) error { return s.client.Close(ctx) }

func (s *Store) EnsureSchema(ctx context.Context) error {
	return neo4j.EnsureConstraints(ctx, s.client, []string{
		`CREATE CONSTRAINT engage_tool_run_id IF NOT EXISTS FOR (n:EngageToolRun) REQUIRE n.id IS UNIQUE`,
		`CREATE CONSTRAINT engage_target_name IF NOT EXISTS FOR (n:EngageTarget) REQUIRE n.name IS UNIQUE`,
		`CREATE CONSTRAINT engage_finding_id IF NOT EXISTS FOR (n:EngageFinding) REQUIRE n.id IS UNIQUE`,
	})
}

// UpsertToolRun persists an engage tool execution and links it to a target host.
func (s *Store) UpsertToolRun(ctx context.Context, id, tool, target, subject string, success bool, at string) error {
	targetName := normalizeTarget(target)
	params := map[string]any{
		"id": id, "tool": tool, "target": target, "subject": subject,
		"success": success, "at": at, "targetName": targetName,
	}
	q := `
MERGE (r:EngageToolRun {id: $id})
SET r.tool = $tool,
    r.target = $target,
    r.subject = $subject,
    r.success = $success,
    r.at = $at
WITH r
MERGE (t:EngageTarget {name: $targetName})
MERGE (t)-[:ENGAGE_RAN]->(r)
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

// UpsertFinding persists a scan finding linked to target and optional tool run.
func (s *Store) UpsertFinding(ctx context.Context, id, tool, target, title, severity, description string) error {
	targetName := normalizeTarget(target)
	params := map[string]any{
		"id": id, "tool": tool, "target": target, "title": title,
		"severity": severity, "description": description, "targetName": targetName,
	}
	cves := extractCVEs(title, description)
	params["cves"] = cves
	q := `
MERGE (f:EngageFinding {id: $id})
SET f.tool = $tool,
    f.target = $target,
    f.title = $title,
    f.severity = $severity,
    f.description = $description
WITH f
MERGE (t:EngageTarget {name: $targetName})
MERGE (t)-[:ENGAGE_FOUND]->(f)
WITH f
UNWIND $cves AS cveId
OPTIONAL MATCH (v:Vulnerability)
WHERE (exists(v.cve) AND toLower(v.cve) = toLower(cveId))
   OR (exists(v.id) AND toLower(v.id) = toLower(cveId))
FOREACH (_ IN CASE WHEN v IS NOT NULL THEN [1] ELSE [] END | MERGE (f)-[:MAY_RELATE_TO]->(v))
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

func normalizeTarget(target string) string {
	t := strings.TrimSpace(target)
	t = strings.TrimPrefix(t, "https://")
	t = strings.TrimPrefix(t, "http://")
	if i := strings.Index(t, "/"); i >= 0 {
		t = t[:i]
	}
	if t == "" {
		return "unknown"
	}
	return t
}
