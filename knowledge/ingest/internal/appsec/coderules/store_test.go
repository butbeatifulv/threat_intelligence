package neo4jstore

import (
	"strings"
	"testing"
)

func TestStableID_deterministic(t *testing.T) {
	a := StableID("semgrep", "path/to/rule.yml")
	b := StableID("semgrep", "path/to/rule.yml")
	if a != b || a == "" {
		t.Fatalf("got %q %q", a, b)
	}
}

func TestClip_truncates(t *testing.T) {
	long := strings.Repeat("a", 30)
	got := clip(long, 10)
	if len(got) != 13 || got[len(got)-3:] != "…" {
		t.Fatalf("got %q len %d", got, len(got))
	}
}
