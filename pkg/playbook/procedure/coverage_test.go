package procedure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/pkg/playbook/domain"
	pbindex "github.com/butbeautifulv/veil/pkg/playbook/index"
	"github.com/butbeautifulv/veil/pkg/playbook/testutil"
)

func TestOpen_errors(t *testing.T) {
	defer pbindex.SetRepoGetwd(func() (string, error) { return "", os.ErrInvalid })()
	t.Setenv(pbindex.EnvRepoRoot, "")
	if _, err := Open(""); err == nil {
		t.Fatal("expected RepoRoot error")
	}

	root := testutil.SetRepoRoot(t)
	bad := filepath.Join(root, "pkg/playbook/testdata/bad-procedures.json")
	if err := os.WriteFile(bad, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(bad); err == nil {
		t.Fatal("expected decode error")
	}
	if _, err := Open(filepath.Join(root, "missing-procedures.json")); err == nil {
		t.Fatal("expected read error")
	}
}

func TestGetSpec_errorsAndCatalogFallback(t *testing.T) {
	testutil.SetRepoRoot(t)
	cat, err := Open(fixtureProceduresIndex)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cat.GetSpec("missing-id"); err == nil {
		t.Fatal("expected unknown id")
	}
	root, _ := pbindex.RepoRoot()
	sum := domain.ProcedureSummary{
		ID:           "fixture-catalog-fallback",
		Subdomain:    "test",
		CatalogTools: []string{"fallback_tool"},
		CorpusPath:   "pkg/playbook/testdata/skills/fixture-catalog-fallback/SKILL.md",
	}
	cat.byID[sum.ID] = sum
	skillDir := filepath.Join(root, filepath.Dir(sum.CorpusPath))
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	raw := `---
name: fallback
---
## When to use
- Plain case

## Workflow
### Step 1: Manual only
No tools named here.

`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	spec, err := cat.GetSpec(sum.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.CatalogTools) != 1 || spec.CatalogTools[0] != "fallback_tool" {
		t.Fatalf("catalog fallback: %v", spec.CatalogTools)
	}
	cat.byID["fixture-bad-read"] = domain.ProcedureSummary{
		ID:         "fixture-bad-read",
		CorpusPath: "pkg/playbook/testdata/skills/does-not-exist/SKILL.md",
	}
	if _, err := cat.GetSpec("fixture-bad-read"); err == nil {
		t.Fatal("expected read error")
	}
}

func TestParseSkillMarkdown_detectionWorkflow(t *testing.T) {
	raw := `---
name: detect
---
## Detection workflow
### Step 1: Observe
Watch logs.

`
	spec := ParseSkillMarkdown("detect", "threat-hunting", nil, nil, raw)
	if len(spec.Steps) != 1 || spec.Steps[0].Title != "Observe" {
		t.Fatalf("steps: %+v", spec.Steps)
	}
}

func TestSplitSections_shortLoc(t *testing.T) {
	old := sectionLocMin
	defer func() { sectionLocMin = old }()
	sectionLocMin = 99
	body := "## Workflow\nline1\n"
	if len(splitSections(body)) != 0 {
		t.Fatal("expected skip of short loc")
	}
}

func TestExtractSteps_shortLoc(t *testing.T) {
	old := stepLocMin
	defer func() { stepLocMin = old }()
	stepLocMin = 99
	if len(extractSteps("### Step 1: X\nbody\n")) != 0 {
		t.Fatal("expected skip of short loc")
	}
}

func TestExtractSteps_truncate(t *testing.T) {
	long := "### Step 1: Big\n" + strings.Repeat("z", 2500) + "\n"
	steps := extractSteps(long)
	if len(steps) != 1 || len(steps[0].Body) != 2000 {
		t.Fatalf("truncate: len=%d", len(steps[0].Body))
	}
}
