package neo4jstore

import "testing"

// Ingest writes (EngageFinding)-[:MAY_RELATE_TO]->(Vulnerability); read path is
// knowledge/connector/query EngageTargetContext and GET /v1/categories/engage/context (Phase 16).

func TestExtractCVEs_dedup(t *testing.T) {
	got := extractCVEs("CVE-2024-1234 in title", "also CVE-2024-1234 and CVE-2023-99999")
	if len(got) != 2 {
		t.Fatalf("len %d: %v", len(got), got)
	}
	if got[0] != "CVE-2024-1234" || got[1] != "CVE-2023-99999" {
		t.Fatalf("%v", got)
	}
}
