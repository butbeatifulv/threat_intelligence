package neo4jstore

import (
	"context"
	"fmt"
	"time"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/butbeautifulv/veil/knowledge/connector/neo4j"
	"github.com/butbeautifulv/veil/pkg/vuln/domain"
	"github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/vuln/repository"
)

type Store struct {
	client *neo4j.Client
}

var _ repository.VulnerabilityRepository = (*Store)(nil)

type Config = neo4j.Config

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
		`CREATE CONSTRAINT vuln_cve IF NOT EXISTS FOR (n:Vulnerability) REQUIRE n.cve IS UNIQUE`,
		`CREATE CONSTRAINT cwe_id IF NOT EXISTS FOR (n:CWE) REQUIRE n.id IS UNIQUE`,
		`CREATE CONSTRAINT cpe_uri IF NOT EXISTS FOR (n:CPE) REQUIRE n.uri IS UNIQUE`,
		`CREATE CONSTRAINT exploit_id IF NOT EXISTS FOR (n:Exploit) REQUIRE n.id IS UNIQUE`,
	})
}

func (s *Store) Save(ctx context.Context, v *domain.Vulnerability) error {
	return s.Upsert(ctx, v)
}

func (s *Store) Upsert(ctx context.Context, v *domain.Vulnerability) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	params := map[string]any{
		"cve":        v.CVE,
		"id":         v.ID,
		"summary":    v.Summary,
		"source":     "nvd",
		"updatedAt":  now,
		"cwes":       v.CWE,
		"cpes":       cpeURIs(v.CPEs),
		"cvss_base":  any(nil),
		"cvss_vec":   any(nil),
		"cvss_ver":   any(nil),
		"markdown":   "# " + v.CVE + "\n\n" + v.Summary,
	}
	if v.CVSS != nil {
		params["cvss_base"] = v.CVSS.Base
		params["cvss_vec"] = v.CVSS.Vector
		params["cvss_ver"] = v.CVSS.Version
	}

	q := `
MERGE (v:Vulnerability {cve: $cve})
SET v.id = $id,
    v.summary = $summary,
    v.markdown = $markdown,
    v.source = $source,
    v.updatedAt = $updatedAt
FOREACH (_ IN CASE WHEN $cvss_base IS NULL THEN [] ELSE [1] END |
  SET v.cvss_base = $cvss_base,
      v.cvss_vector = $cvss_vec,
      v.cvss_version = $cvss_ver
)

WITH v
UNWIND (CASE WHEN $cwes IS NULL THEN [] ELSE $cwes END) AS cweId
WITH v, cweId WHERE cweId IS NOT NULL AND cweId <> ""
MERGE (cwe:CWE {id: cweId})
MERGE (v)-[:HAS_CWE]->(cwe)

WITH v
UNWIND (CASE WHEN $cpes IS NULL THEN [] ELSE $cpes END) AS cpeUri
WITH v, cpeUri WHERE cpeUri IS NOT NULL AND cpeUri <> ""
MERGE (cpe:CPE {uri: cpeUri})
MERGE (v)-[:AFFECTS]->(cpe)
`

	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

func cpeURIs(cpes []domain.CPE) []string {
	if len(cpes) == 0 {
		return nil
	}
	out := make([]string, 0, len(cpes))
	for _, c := range cpes {
		if c.URI != "" {
			out = append(out, c.URI)
		}
	}
	return out
}

func stableExploitID(source, refID string) string {
	h := source + ":" + refID
	var x uint64 = 14695981039346656037
	for _, b := range []byte(h) {
		x ^= uint64(b)
		x *= 1099511628211
	}
	return fmt.Sprintf("exploit:%s:%016x", source, x)
}

func (s *Store) MergeExploitForCVE(ctx context.Context, cve string, ref domain.ExploitRef) error {
	if cve == "" || ref.RefID == "" {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	exploitID := stableExploitID(ref.Source, ref.RefID)
	params := map[string]any{
		"cve":       cve,
		"exploitID": exploitID,
		"source":    ref.Source,
		"refId":     ref.RefID,
		"url":       ref.URL,
		"now":       now,
	}
	q := `
MATCH (v:Vulnerability {cve: $cve})
MERGE (e:Exploit {id: $exploitID})
SET e.source = $source,
    e.refId = $refId,
    e.url = $url,
    e.updatedAt = $now
MERGE (v)-[r:HAS_EXPLOIT]->(e)
SET r.updatedAt = $now
`
	return s.client.ExecWrite(ctx, func(tx driver.ManagedTransaction) error {
		_, err := tx.Run(ctx, q, params)
		return err
	})
}

