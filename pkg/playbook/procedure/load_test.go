package procedure

import (
	"path/filepath"
	"testing"

	"github.com/butbeautifulv/veil/pkg/playbook/testutil"
)

const fixtureProceduresIndex = "pkg/playbook/testdata/procedures-index.json"

func TestOpen_fixtureIndex(t *testing.T) {
	testutil.SetRepoRoot(t)
	cat, err := Open(fixtureProceduresIndex)
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.file.Procedures) != 2 {
		t.Fatalf("procedures: got %d want 2", len(cat.file.Procedures))
	}
	if cat.file.SkillCount != 2 {
		t.Fatalf("skill_count: %d", cat.file.SkillCount)
	}
}

func TestGetSummary_fixture(t *testing.T) {
	testutil.SetRepoRoot(t)
	cat, err := Open(fixtureProceduresIndex)
	if err != nil {
		t.Fatal(err)
	}
	sum, ok := cat.GetSummary("fixture-forensics")
	if !ok {
		t.Fatal("expected fixture-forensics")
	}
	if sum.Subdomain != "digital-forensics" {
		t.Fatalf("subdomain: %q", sum.Subdomain)
	}
	if _, ok := cat.GetSummary("missing-id"); ok {
		t.Fatal("expected unknown id")
	}
}

func TestGetSpec_fixture(t *testing.T) {
	testutil.SetRepoRoot(t)
	cat, err := Open(fixtureProceduresIndex)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := cat.GetSpec("fixture-forensics")
	if err != nil {
		t.Fatal(err)
	}
	if spec.ID != "fixture-forensics" {
		t.Fatalf("id: %q", spec.ID)
	}
	if len(spec.WhenToUse) < 2 {
		t.Fatalf("when_to_use: %d", len(spec.WhenToUse))
	}
	if len(spec.Prerequisites) != 1 {
		t.Fatalf("prerequisites: %d", len(spec.Prerequisites))
	}
	if len(spec.Steps) < 2 {
		t.Fatalf("steps: %d", len(spec.Steps))
	}
	if len(spec.CatalogTools) == 0 {
		t.Fatal("expected catalog tools from mentions or summary")
	}
}

func TestBySubdomain_fixture(t *testing.T) {
	testutil.SetRepoRoot(t)
	cat, err := Open(fixtureProceduresIndex)
	if err != nil {
		t.Fatal(err)
	}
	forensics := cat.BySubdomain("digital-forensics")
	if len(forensics) != 1 || forensics[0].ID != "fixture-forensics" {
		t.Fatalf("forensics: %+v", forensics)
	}
	hunt := cat.BySubdomain("THREAT-HUNTING")
	if len(hunt) != 1 || hunt[0].ID != "fixture-hunt" {
		t.Fatalf("hunt: %+v", hunt)
	}
	if len(cat.BySubdomain("nonexistent")) != 0 {
		t.Fatal("expected empty for unknown subdomain")
	}
}

func TestOpen_defaultIndex_integration(t *testing.T) {
	root := testutil.SetRepoRoot(t)
	idx := filepath.Join(root, DefaultProceduresRel)
	if _, err := Open(""); err != nil {
		t.Skipf("default procedures index: %v", err)
	}
	cat, err := Open("")
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.file.Procedures) < 100 {
		t.Fatalf("procedures: %d at %s", len(cat.file.Procedures), idx)
	}
}
