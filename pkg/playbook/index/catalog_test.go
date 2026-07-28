package index

import (
	"testing"

	"github.com/butbeautifulv/veil/pkg/playbook/testutil"
)

const fixtureSkillsIndex = "pkg/playbook/testdata/cyber-skills.json"

func openFixture(t *testing.T) *Catalog {
	t.Helper()
	testutil.SetRepoRoot(t)
	cat, err := Open(fixtureSkillsIndex)
	if err != nil {
		t.Fatal(err)
	}
	return cat
}

func TestReadBody_fixture(t *testing.T) {
	cat := openFixture(t)
	detail, err := cat.ReadBody("fixture-forensics")
	if err != nil {
		t.Fatal(err)
	}
	if detail.ID != "fixture-forensics" {
		t.Fatalf("id: %q", detail.ID)
	}
	if len(detail.Body) == 0 {
		t.Fatal("expected body")
	}
	if _, err := cat.ReadBody("no-such-skill"); err == nil {
		t.Fatal("expected error for unknown skill")
	}
}

func TestSearch_fixture(t *testing.T) {
	cat := openFixture(t)
	hits := cat.Search("disk imaging forensic", "", 5)
	if len(hits) == 0 || hits[0].ID != "fixture-forensics" {
		t.Fatalf("forensics search: %+v", hits)
	}
	filtered := cat.Search("powershell", "threat-hunting", 5)
	if len(filtered) != 1 || filtered[0].ID != "fixture-hunt" {
		t.Fatalf("subdomain filter: %+v", filtered)
	}
	if len(cat.Search("nomatch-xyz-abc", "", 5)) != 0 {
		t.Fatal("expected no hits")
	}
}

func TestByTechnique_fixture(t *testing.T) {
	cat := openFixture(t)
	hits := cat.ByTechnique("T1059.001")
	if len(hits) != 1 || hits[0].ID != "fixture-hunt" {
		t.Fatalf("T1059.001: %+v", hits)
	}
	if len(cat.ByTechnique("t1005")) != 1 {
		t.Fatalf("T1005: %+v", cat.ByTechnique("t1005"))
	}
	if len(cat.ByTechnique("T9999")) != 0 {
		t.Fatal("expected no skills")
	}
}
