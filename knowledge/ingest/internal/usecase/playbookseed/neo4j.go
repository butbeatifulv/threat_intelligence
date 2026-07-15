package playbookseed

import (
	"context"
	"strings"

	graphneo4j "github.com/butbeautifulv/veil/knowledge/connector/neo4j"
	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Neo4jRepository implements Repository via the graph connector client.
type Neo4jRepository struct {
	Client *graphneo4j.Client
}

func (r *Neo4jRepository) EnsureConstraints(ctx context.Context) error {
	return graphneo4j.EnsureConstraints(ctx, r.Client, []string{
		`CREATE CONSTRAINT cyber_skill_id IF NOT EXISTS FOR (n:CyberSkill) REQUIRE n.id IS UNIQUE`,
	})
}

func (r *Neo4jRepository) UpsertSkill(ctx context.Context, s SkillSeed) error {
	return r.Client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, `
MERGE (sk:CyberSkill {id: $id})
SET sk.title = $title, sk.subdomain = $subdomain, sk.source = 'anthropic-cyber-skills',
    sk.stepCount = $stepCount, sk.hasStructured = $hasStructured, sk.updatedAt = datetime()`,
			map[string]any{
				"id": s.ID, "title": s.Name, "subdomain": s.Subdomain,
				"stepCount": s.StepCount, "hasStructured": s.HasStructured,
			})
		return err
	})
}

func (r *Neo4jRepository) LinkAttackTechnique(ctx context.Context, attackID, skillID string) error {
	return r.Client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, `
MATCH (t:AttackTechnique {id: $tid}), (sk:CyberSkill {id: $sid})
MERGE (t)-[r:HAS_PLAYBOOK]->(sk)
SET r.updatedAt = datetime()`,
			map[string]any{"tid": strings.ToUpper(attackID), "sid": skillID})
		return err
	})
}
