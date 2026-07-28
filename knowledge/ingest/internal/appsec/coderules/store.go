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
		`CREATE CONSTRAINT semgrep_rule_id IF NOT EXISTS FOR (n:SemgrepRule) REQUIRE n.id IS UNIQUE`,
		`CREATE CONSTRAINT codeql_rule_id IF NOT EXISTS FOR (n:CodeQLRule) REQUIRE n.id IS UNIQUE`,
		`CREATE CONSTRAINT cwe_id IF NOT EXISTS FOR (n:CWE) REQUIRE n.id IS UNIQUE`,
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

func clip(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func (s *Store) UpsertCWECatalog(ctx context.Context, cweID, name, description, status string) error {
	if cweID == "" {
		return nil
	}
	if !strings.HasPrefix(strings.ToUpper(cweID), "CWE-") {
		cweID = "CWE-" + strings.TrimPrefix(strings.TrimPrefix(cweID, "cwe-"), "CWE-")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	params := map[string]any{
		"id":          cweID,
		"name":        clip(name, 2000),
		"description": clip(description, 8000),
		"status":      clip(status, 256),
		"updatedAt":   now,
	}
	q := `
MERGE (n:CWE {id: $id})
SET n.name = CASE WHEN $name = "" THEN n.name ELSE $name END,
    n.description = CASE WHEN $description = "" THEN n.description ELSE $description END,
    n.status = CASE WHEN $status = "" THEN n.status ELSE $status END,
    n.catalogSource = "mitre-cwe",
    n.updatedAt = $updatedAt
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

func (s *Store) UpsertSemgrepRule(ctx context.Context, id, path, title, lang, markdown string) error {
	if id == "" {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	params := map[string]any{
		"id": id, "path": path, "title": title, "lang": lang,
		"markdown": markdown, "source": "semgrep-rules", "updatedAt": now,
	}
	q := `
MERGE (n:SemgrepRule {id: $id})
SET n.path = $path,
    n.title = $title,
    n.language = $lang,
    n.markdown = $markdown,
    n.source = $source,
    n.updatedAt = $updatedAt
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

// LinkSemgrepRuleToCWE creates MAPS_TO_CWE from an ingested Semgrep rule to a CWE id (CWE-nnn).
func (s *Store) LinkSemgrepRuleToCWE(ctx context.Context, ruleID, cweID string) error {
	ruleID = strings.TrimSpace(ruleID)
	cweID = normalizeCWEID(cweID)
	if ruleID == "" || cweID == "" {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	params := map[string]any{"rid": ruleID, "cwe": cweID, "ts": now}
	q := `
MATCH (r:SemgrepRule {id: $rid})
MERGE (c:CWE {id: $cwe})
MERGE (r)-[x:MAPS_TO_CWE]->(c)
SET x.updatedAt = $ts
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

// LinkCodeQLRuleToCWE links a CodeQL query node to CWE when metadata references it.
func (s *Store) LinkCodeQLRuleToCWE(ctx context.Context, ruleID, cweID string) error {
	ruleID = strings.TrimSpace(ruleID)
	cweID = normalizeCWEID(cweID)
	if ruleID == "" || cweID == "" {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	params := map[string]any{"rid": ruleID, "cwe": cweID, "ts": now}
	q := `
MATCH (r:CodeQLRule {id: $rid})
MERGE (c:CWE {id: $cwe})
MERGE (r)-[x:MAPS_TO_CWE]->(c)
SET x.updatedAt = $ts
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

func normalizeCWEID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	raw = strings.TrimPrefix(strings.ToUpper(raw), "CWE-")
	// "79: ..." or "CWE-79: ..."
	if i := strings.IndexAny(raw, " :;,\t"); i > 0 {
		raw = raw[:i]
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	return "CWE-" + strings.TrimPrefix(raw, "CWE-")
}

func (s *Store) UpsertCodeQLRule(ctx context.Context, id, path, name, lang, markdown string) error {
	if id == "" {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	params := map[string]any{
		"id": id, "path": path, "name": name, "lang": lang,
		"markdown": markdown, "source": "github-codeql", "updatedAt": now,
	}
	q := `
MERGE (n:CodeQLRule {id: $id})
SET n.path = $path,
    n.name = $name,
    n.language = $lang,
    n.markdown = $markdown,
    n.source = $source,
    n.updatedAt = $updatedAt
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}
