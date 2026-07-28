package query

import (
	"strings"
	"testing"
)

func TestSafeLabel(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"Vulnerability", "Vulnerability"},
		{"IOC-2024", "IOC2024"},
		{"'; DROP", "DROP"},
		{"", "Node"},
		{"___", "___"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := safeLabel(tt.in); got != tt.want {
				t.Fatalf("safeLabel(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestBuildNodesByKindCypher(t *testing.T) {
	cy := buildNodesByKindCypher("Vulnerability")
	if !strings.Contains(cy, "MATCH (n:Vulnerability)") {
		t.Fatalf("label match: %s", cy)
	}
	if !strings.Contains(cy, "SKIP $offset LIMIT $limit") {
		t.Fatal("pagination missing")
	}
	cyBad := buildNodesByKindCypher("bad;label")
	if strings.Contains(cyBad, "bad;") {
		t.Fatalf("injection not stripped: %s", cyBad)
	}
}

func TestBuildGetNodeCypher_includesEngageTarget(t *testing.T) {
	cy := buildGetNodeCypher()
	if !strings.Contains(cy, "EngageTarget") {
		t.Fatal("missing EngageTarget name match")
	}
	if !strings.Contains(cy, nodeMatchByID) {
		t.Fatal("missing nodeMatchByID")
	}
}

func TestBuildSearchCypher(t *testing.T) {
	tests := []struct {
		kind    string
		wantLbl bool
	}{
		{"", false},
		{"IOC", true},
	}
	for _, tt := range tests {
		cy := buildSearchCypher(tt.kind)
		if !strings.Contains(cy, nodeTextSearchPredicate) {
			t.Fatal("missing text predicate")
		}
		if tt.wantLbl && !strings.Contains(cy, "MATCH (n:IOC)") {
			t.Fatalf("kind label: %s", cy)
		}
		if !tt.wantLbl && strings.Contains(cy, "MATCH (n:") && !strings.Contains(cy, "MATCH (n)\n") {
			// unlabeled search uses MATCH (n) without label suffix
			if strings.Contains(cy, "MATCH (n:IOC)") {
				t.Fatalf("unexpected label: %s", cy)
			}
		}
	}
}

func TestBuildSearchInCategoryCypher(t *testing.T) {
	cy := buildSearchInCategoryCypher("EngageFinding")
	if !strings.Contains(cy, "MATCH (n:EngageFinding)") {
		t.Fatalf("kind: %s", cy)
	}
	if !strings.Contains(cy, "l IN $allowed") {
		t.Fatal("category filter missing")
	}
	cyAll := buildSearchInCategoryCypher("")
	if !strings.Contains(cyAll, "any(l IN labels(n) WHERE l IN $allowed)") {
		t.Fatalf("allowlist: %s", cyAll)
	}
}

func TestBuildNeighborsCypher_depth(t *testing.T) {
	cy := buildNeighborsNodesCypher(2)
	if !strings.Contains(cy, "[r*1..2]") {
		t.Fatalf("depth: %s", cy)
	}
	if !strings.Contains(cy, seedMatchByID) {
		t.Fatal("seed match missing")
	}
	edges := buildNeighborsEdgesCypher(3)
	if !strings.Contains(edges, "[r*1..3]") {
		t.Fatalf("edges depth: %s", edges)
	}
}

func TestClampHelpers(t *testing.T) {
	if clampLimit(0, 50, 2000) != 50 {
		t.Fatal("default limit")
	}
	if clampLimit(9999, 50, 2000) != 50 {
		t.Fatal("max limit")
	}
	if clampLimit(10, 50, 2000) != 10 {
		t.Fatal("valid limit")
	}
	if clampOffset(-1) != 0 {
		t.Fatal("offset")
	}
	if clampDepth(0) != 1 || clampDepth(9) != 1 || clampDepth(2) != 2 {
		t.Fatal("depth")
	}
}
