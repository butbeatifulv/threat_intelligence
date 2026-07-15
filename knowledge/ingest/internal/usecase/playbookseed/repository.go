package playbookseed

import "context"

// SkillSeed is one cyber skill row to MERGE into the graph.
type SkillSeed struct {
	ID            string
	Name          string
	Subdomain     string
	StepCount     int
	HasStructured bool
	AttackIDs     []string
}

// Repository persists playbook seed data (Neo4j adapter implements this port).
type Repository interface {
	EnsureConstraints(ctx context.Context) error
	UpsertSkill(ctx context.Context, s SkillSeed) error
	LinkAttackTechnique(ctx context.Context, attackID, skillID string) error
}
