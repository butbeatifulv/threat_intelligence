// Seed CyberSkill nodes and HAS_PLAYBOOK edges from docs/skills-index/cyber-skills.json.
package main

import (
	"context"
	"log"
	"os"
	"time"

	graphneo4j "github.com/butbeautifulv/veil/knowledge/connector/neo4j"
	"github.com/butbeautifulv/veil/knowledge/ingest/internal/usecase/playbookseed"
	pbindex "github.com/butbeautifulv/veil/pkg/playbook/index"
	pbprocedure "github.com/butbeautifulv/veil/pkg/playbook/procedure"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cat, err := pbindex.Open("")
	if err != nil {
		log.Fatal(err)
	}
	procCat, _ := pbprocedure.Open("")
	procByID := map[string]int{}
	if procCat != nil {
		for _, p := range procCat.Meta().Procedures {
			procByID[p.ID] = p.StepCount
		}
	}
	cfg := graphneo4j.Config{
		URI:      envOr("NEO4J_URI", "neo4j://localhost:7687"),
		Username: envOr("NEO4J_USERNAME", "neo4j"),
		Password: envOr("NEO4J_PASSWORD", "neo4jpassword"),
		Database: envOr("NEO4J_DATABASE", "neo4j"),
	}
	client, err := graphneo4j.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = client.Close(ctx) }()

	meta := cat.Meta()
	seeds := make([]playbookseed.SkillSeed, 0, len(meta.Skills))
	for _, s := range meta.Skills {
		steps := procByID[s.ID]
		seeds = append(seeds, playbookseed.SkillSeed{
			ID: s.ID, Name: s.Name, Subdomain: s.Subdomain,
			StepCount: steps, HasStructured: steps > 0,
			AttackIDs: s.AttackIDs,
		})
	}
	repo := &playbookseed.Neo4jRepository{Client: client}
	skills, linked, err := playbookseed.Run(ctx, repo, playbookseed.CatalogMeta{Skills: seeds}, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("seeded %d skills, %d HAS_PLAYBOOK edges", skills, linked)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
