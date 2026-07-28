package neo4jstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/butbeautifulv/veil/pkg/ti/domain"
	"github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/ti/repository"

	graphneo4j "github.com/butbeautifulv/veil/knowledge/connector/neo4j"
)

type Store struct {
	client *graphneo4j.Client
}

var _ repository.GraphRepository = (*Store)(nil)

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
	queries := []string{
		`DROP CONSTRAINT actor_name IF EXISTS`,
		`CREATE CONSTRAINT ioc_id IF NOT EXISTS FOR (n:IOC) REQUIRE n.id IS UNIQUE`,
		`CREATE CONSTRAINT campaign_id IF NOT EXISTS FOR (n:Campaign) REQUIRE n.id IS UNIQUE`,
		`CREATE CONSTRAINT cluster_id IF NOT EXISTS FOR (n:Cluster) REQUIRE n.id IS UNIQUE`,
		`CREATE CONSTRAINT actor_id IF NOT EXISTS FOR (n:Actor) REQUIRE n.id IS UNIQUE`,
		`CREATE CONSTRAINT report_id IF NOT EXISTS FOR (n:Report) REQUIRE n.id IS UNIQUE`,
	}
	return graphneo4j.EnsureConstraints(ctx, s.client, queries)
}

func (s *Store) UpsertIOC(ctx context.Context, id string, i domain.IOC) error {
	eventTime := time.Now().UTC().Format(time.RFC3339Nano)
	sources := i.Sources
	if sources == nil {
		sources = []string{}
	}
	params := map[string]any{
		"id":          id,
		"type":        string(i.Type),
		"value":       i.Value,
		"source":      i.Source,
		"sources":     sources,
		"confidence":  i.Confidence,
		"tags":        i.Tags,
		"updatedAt":   eventTime,
		"eventTime":   eventTime,
	}
	q := `
MERGE (n:IOC {id: $id})
ON CREATE SET n.firstSeen = $eventTime, n.lastSeen = $eventTime
ON MATCH SET n.lastSeen = CASE WHEN $eventTime > coalesce(n.lastSeen, '') THEN $eventTime ELSE n.lastSeen END,
             n.firstSeen = CASE WHEN $eventTime < n.firstSeen THEN $eventTime ELSE n.firstSeen END
SET n.type = $type,
    n.value = $value,
    n.source = CASE WHEN $source = "" THEN n.source ELSE $source END,
    n.updatedAt = $updatedAt,
    n.sources = apoc.coll.sort(apoc.coll.toSet(coalesce(n.sources, []) + $sources))
FOREACH (_ IN CASE WHEN $confidence IS NULL THEN [] ELSE [1] END | SET n.confidence = $confidence)
FOREACH (_ IN CASE WHEN $tags IS NULL OR size($tags) = 0 THEN [] ELSE [1] END | SET n.tags = $tags)
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

func (s *Store) UpsertCampaign(ctx context.Context, c domain.Campaign) error {
	params := map[string]any{
		"id":        c.ID,
		"name":      c.Name,
		"summary":   c.Summary,
		"source":    c.Source,
		"updatedAt": time.Now().UTC().Format(time.RFC3339Nano),
	}
	q := `
MERGE (n:Campaign {id: $id})
SET n.name = $name,
    n.summary = $summary,
    n.source = CASE WHEN $source = "" THEN n.source ELSE $source END,
    n.updatedAt = $updatedAt
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

func (s *Store) UpsertCluster(ctx context.Context, cl domain.Cluster) error {
	params := map[string]any{
		"id":          cl.ID,
		"name":        cl.Name,
		"description": cl.Description,
		"source":      cl.Source,
		"updatedAt":   time.Now().UTC().Format(time.RFC3339Nano),
	}
	q := `
MERGE (n:Cluster {id: $id})
SET n.name = $name,
    n.description = $description,
    n.source = CASE WHEN $source = "" THEN n.source ELSE $source END,
    n.updatedAt = $updatedAt
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

func (s *Store) LinkCampaignIOC(ctx context.Context, campaignID, iocID string) error {
	params := map[string]any{
		"campaignID": campaignID,
		"iocID":      iocID,
		"ts":         time.Now().UTC().Format(time.RFC3339Nano),
	}
	q := `
MATCH (c:Campaign {id: $campaignID})
MATCH (i:IOC {id: $iocID})
MERGE (c)-[r:INDICATOR]->(i)
SET r.updatedAt = $ts
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

func (s *Store) LinkClusterCampaign(ctx context.Context, clusterID string, campaignID string) error {
	params := map[string]any{
		"clusterID":  clusterID,
		"campaignID": campaignID,
		"ts":         time.Now().UTC().Format(time.RFC3339Nano),
	}
	q := `
MATCH (cl:Cluster {id: $clusterID})
MATCH (c:Campaign {id: $campaignID})
MERGE (cl)-[r:HAS_CAMPAIGN]->(c)
SET r.updatedAt = $ts
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

func (s *Store) LinkCampaignActor(ctx context.Context, campaignID, actorID, actorName string) error {
	params := map[string]any{
		"campaignID": campaignID,
		"actorId":    actorID,
		"name":       actorName,
		"ts":         time.Now().UTC().Format(time.RFC3339Nano),
	}
	q := `
MERGE (a:Actor {id: $actorId})
SET a.name = CASE WHEN coalesce(a.name,"") = "" THEN $name ELSE a.name END,
    a.updatedAt = $ts
WITH a
MATCH (c:Campaign {id: $campaignID})
MERGE (c)-[r:ATTRIBUTED_TO]->(a)
SET r.updatedAt = $ts
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

func (s *Store) UpsertActor(ctx context.Context, a domain.Actor) error {
	id := strings.TrimSpace(a.ID)
	if id == "" {
		return fmt.Errorf("ti actor: empty id (payload must be normalized by NED)")
	}
	aliasesJSON, _ := json.Marshal(a.Aliases)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	md := fmt.Sprintf("# %s\n\n%s", a.Name, a.Description)
	params := map[string]any{
		"id":          id,
		"name":        a.Name,
		"aliases":     string(aliasesJSON),
		"description": a.Description,
		"source":      a.Source,
		"markdown":    md,
		"updatedAt":   now,
	}
	q := `
MERGE (n:Actor {id: $id})
SET n.name = $name,
    n.aliases = $aliases,
    n.description = CASE WHEN $description = "" THEN n.description ELSE $description END,
    n.source = CASE WHEN $source = "" THEN n.source ELSE $source END,
    n.markdown = $markdown,
    n.updatedAt = $updatedAt
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

func (s *Store) UpsertReport(ctx context.Context, r domain.Report) error {
	id := strings.TrimSpace(r.ID)
	if id == "" {
		return fmt.Errorf("ti report: empty id (payload must be normalized by NED)")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	md := fmt.Sprintf("# %s\n\n**Provider:** %s  \n**Link:** %s  \n\n%s", r.Title, r.Provider, r.Link, r.BodyMarkdown)
	params := map[string]any{
		"id":           id,
		"title":        r.Title,
		"provider":     r.Provider,
		"link":         r.Link,
		"publishedAt":  r.PublishedAt,
		"bodyMarkdown": r.BodyMarkdown,
		"source":       r.Source,
		"markdown":     md,
		"updatedAt":    now,
	}
	q := `
MERGE (n:Report {id: $id})
SET n.title = $title,
    n.provider = $provider,
    n.link = $link,
    n.publishedAt = CASE WHEN $publishedAt = "" THEN n.publishedAt ELSE $publishedAt END,
    n.bodyMarkdown = CASE WHEN $bodyMarkdown = "" THEN n.bodyMarkdown ELSE $bodyMarkdown END,
    n.markdown = $markdown,
    n.source = CASE WHEN $source = "" THEN n.source ELSE $source END,
    n.updatedAt = $updatedAt
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

func (s *Store) LinkReportMentionsIOC(ctx context.Context, reportID, iocID string) error {
	if reportID == "" {
		return nil
	}
	params := map[string]any{
		"reportID": reportID,
		"iocID":    iocID,
		"ts":       time.Now().UTC().Format(time.RFC3339Nano),
	}
	q := `
MATCH (rep:Report {id: $reportID})
MATCH (ioc:IOC {id: $iocID})
MERGE (rep)-[r:MENTIONS]->(ioc)
SET r.updatedAt = $ts
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

func (s *Store) UpsertKEVVulnerability(ctx context.Context, cve, vendor, product, summary, dateAdded string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	md := fmt.Sprintf("# %s (KEV)\n\n**Vendor:** %s\n\n**Product:** %s\n\n**Added:** %s\n\n%s", cve, vendor, product, dateAdded, summary)
	params := map[string]any{
		"cve":       cve,
		"vendor":    vendor,
		"product":   product,
		"summary":   summary,
		"dateAdded": dateAdded,
		"updatedAt": now,
		"source":    "cisa-kev",
		"markdown":  md,
	}
	q := `
MERGE (v:Vulnerability {cve: $cve})
SET v.kev = true,
    v.kevVendor = $vendor,
    v.kevProduct = $product,
    v.kevDateAdded = $dateAdded,
    v.source = CASE WHEN coalesce(v.source,"") = "" THEN $source ELSE v.source END,
    v.summary = CASE WHEN coalesce(v.summary,"") = "" THEN $summary ELSE v.summary END,
    v.markdown = $markdown,
    v.updatedAt = $updatedAt
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

