package framework

import (
	"testing"

	"github.com/butbeautifulv/veil/pkg/playbook/testutil"
)

func TestLoadSubdomains(t *testing.T) {
	testutil.SetRepoRoot(t)
	subs, err := LoadSubdomains()
	if err != nil {
		t.Fatal(err)
	}
	if len(subs) < 20 {
		t.Fatalf("subdomains: %d", len(subs))
	}
	foundKnown := false
	for _, s := range subs {
		if s.ID == "penetration-testing" && s.Priority == "P2" && len(s.VeilCats) > 0 {
			foundKnown = true
		}
	}
	if !foundKnown {
		t.Fatal("expected known subdomain entry with veil categories")
	}
}

func TestSkillsForTechnique_T1059(t *testing.T) {
	testutil.SetRepoRoot(t)
	ids, err := SkillsForTechnique("T1059.001")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) == 0 {
		t.Fatal("expected skills for T1059.001")
	}
	idsLower, err := SkillsForTechnique("t1059.001")
	if err != nil {
		t.Fatal(err)
	}
	if len(idsLower) != len(ids) {
		t.Fatalf("case fold: %d vs %d", len(idsLower), len(ids))
	}
}
