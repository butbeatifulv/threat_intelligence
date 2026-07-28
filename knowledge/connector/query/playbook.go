package query

import (
	"context"
	"strings"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// PlaybookSkillRef is a CyberSkill node linked from ATT&CK.
type PlaybookSkillRef struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Subdomain string `json:"subdomain"`
}

// PlaybooksForTechnique returns CyberSkill nodes linked via HAS_PLAYBOOK from AttackTechnique.
func (s *Service) PlaybooksForTechnique(ctx context.Context, techniqueID string) ([]PlaybookSkillRef, error) {
	techniqueID = strings.TrimSpace(techniqueID)
	if techniqueID == "" {
		return nil, nil
	}
	raw, err := s.exec.ExecRead(ctx, func(tx driver.ManagedTransaction) (any, error) {
		q := `
MATCH (t:AttackTechnique {id: $tid})-[:HAS_PLAYBOOK]->(s:CyberSkill)
RETURN s.id AS id, s.title AS title, s.subdomain AS subdomain
ORDER BY s.title
LIMIT 50`
		res, err := tx.Run(ctx, q, map[string]any{"tid": techniqueID})
		if err != nil {
			return nil, err
		}
		var out []PlaybookSkillRef
		for res.Next(ctx) {
			rec := res.Record()
			ref := PlaybookSkillRef{}
			if v, ok := rec.Get("id"); ok {
				ref.ID, _ = v.(string)
			}
			if v, ok := rec.Get("title"); ok {
				ref.Title, _ = v.(string)
			}
			if v, ok := rec.Get("subdomain"); ok {
				ref.Subdomain, _ = v.(string)
			}
			if ref.ID != "" {
				out = append(out, ref)
			}
		}
		return out, res.Err()
	})
	if err != nil {
		return nil, err
	}
	out, _ := raw.([]PlaybookSkillRef)
	return out, nil
}
