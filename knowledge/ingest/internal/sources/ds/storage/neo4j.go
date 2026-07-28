package neo4jstore

import (
	"context"
	"fmt"
	"strings"
	"time"

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
		`CREATE CONSTRAINT sigma_rule_id IF NOT EXISTS FOR (n:SigmaRule) REQUIRE n.id IS UNIQUE`,
		`CREATE CONSTRAINT yara_rule_id IF NOT EXISTS FOR (n:YaraRule) REQUIRE n.id IS UNIQUE`,
		`CREATE CONSTRAINT atomic_test_id IF NOT EXISTS FOR (n:AtomicTest) REQUIRE n.id IS UNIQUE`,
		`CREATE CONSTRAINT caldera_ability_id IF NOT EXISTS FOR (n:CalderaAbility) REQUIRE n.id IS UNIQUE`,
	})
}

func (s *Store) UpsertSigmaRule(ctx context.Context, id, title, level, logProduct, logService, tagsJSON, markdown, source string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	params := map[string]any{
		"id": id, "title": title, "level": level, "logProduct": logProduct, "logService": logService,
		"tags": tagsJSON, "markdown": markdown, "source": source, "updatedAt": now,
	}
	q := `
MERGE (n:SigmaRule {id: $id})
SET n.title = $title,
    n.level = $level,
    n.logsource_product = $logProduct,
    n.logsource_service = $logService,
    n.tags = $tags,
    n.markdown = $markdown,
    n.source = $source,
    n.updatedAt = $updatedAt
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

func (s *Store) UpsertYaraRule(ctx context.Context, id, name, author, tagsJSON, markdown, source string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	params := map[string]any{
		"id": id, "name": name, "author": author, "tags": tagsJSON, "markdown": markdown, "source": source, "updatedAt": now,
	}
	q := `
MERGE (n:YaraRule {id: $id})
SET n.name = $name,
    n.author = $author,
    n.tags = $tags,
    n.markdown = $markdown,
    n.source = $source,
    n.updatedAt = $updatedAt
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

func (s *Store) UpsertAtomicTest(ctx context.Context, id, name, tactic, technique, execName, execCmd, markdown, source string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	params := map[string]any{
		"id": id, "name": name, "tactic": tactic, "technique": technique,
		"execName": execName, "execCmd": execCmd, "markdown": markdown, "source": source, "updatedAt": now,
	}
	q := `
MERGE (n:AtomicTest {id: $id})
SET n.name = $name,
    n.tactic = $tactic,
    n.technique = $technique,
    n.executor_name = $execName,
    n.executor_command = $execCmd,
    n.markdown = $markdown,
    n.source = $source,
    n.updatedAt = $updatedAt
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
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

func (s *Store) UpsertCalderaAbility(ctx context.Context, id, name, tactic, techniqueID, markdown, source string) error {
	if id == "" {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	params := map[string]any{
		"id": id, "name": name, "tactic": tactic, "technique": techniqueID,
		"markdown": markdown, "source": source, "updatedAt": now,
	}
	q := `
MERGE (n:CalderaAbility {id: $id})
SET n.name = $name,
    n.tactic = $tactic,
    n.technique = $technique,
    n.markdown = $markdown,
    n.source = $source,
    n.updatedAt = $updatedAt
`
	if err := s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	}); err != nil {
		return err
	}
	if strings.TrimSpace(techniqueID) == "" {
		return nil
	}
	q2 := `
MATCH (c:CalderaAbility {id: $id})
OPTIONAL MATCH (t:AttackTechnique {id: $technique})
WITH c, t WHERE t IS NOT NULL
MERGE (c)-[r:RELATES_TO_TECHNIQUE]->(t)
SET r.updatedAt = $updatedAt
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q2, map[string]any{"id": id, "technique": techniqueID, "updatedAt": now})
		return err
	})
}
