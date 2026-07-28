package index

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/pkg/playbook/domain"
	"github.com/butbeautifulv/veil/pkg/playbook/testutil"
)

func TestOpen_errorsAndPaths(t *testing.T) {
	defer SetRepoGetwd(func() (string, error) { return "", os.ErrInvalid })()
	if _, err := Open(""); err == nil {
		t.Fatal("expected RepoRoot error")
	}

	root := testutil.SetRepoRoot(t)
	idxDir := filepath.Join(root, "docs/skills-index")
	if err := os.MkdirAll(idxDir, 0o755); err != nil {
		t.Fatal(err)
	}
	badPath := filepath.Join(idxDir, "bad-index.json")
	if err := os.WriteFile(badPath, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(badPath); err == nil {
		t.Fatal("expected decode error")
	}
	if _, err := Open(filepath.Join(root, "no-such-index.json")); err == nil {
		t.Fatal("expected read error")
	}

	good := filepath.Join(root, fixtureSkillsIndex)
	t.Setenv(EnvIndexPath, fixtureSkillsIndex)
	cat, err := Open("")
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.file.Skills) == 0 {
		t.Fatal("expected skills via env path")
	}
	abs, err := Open(good)
	if err != nil {
		t.Fatal(err)
	}
	if len(abs.file.Skills) == 0 {
		t.Fatal("expected skills via abs path")
	}
}

func TestRepoRoot_discoverAndFallback(t *testing.T) {
	root := testutil.SetRepoRoot(t)
	t.Setenv(EnvRepoRoot, "")
	discoverRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(discoverRoot, "versions.env"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(discoverRoot, "go.mod"), []byte("module test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	deep := filepath.Join(discoverRoot, "a", "b")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(deep); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(root) })
	got, err := RepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	if got != discoverRoot {
		t.Fatalf("discover: got %q want %q", got, discoverRoot)
	}

	isolated := t.TempDir()
	if err := os.Chdir(isolated); err != nil {
		t.Fatal(err)
	}
	got, err = RepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	if got != isolated {
		t.Fatalf("fallback wd: got %q want %q", got, isolated)
	}

	defer SetRepoGetwd(func() (string, error) { return "", os.ErrInvalid })()
	if _, err := RepoRoot(); err == nil {
		t.Fatal("expected getwd error")
	}
	t.Setenv(EnvRepoRoot, "  /custom/root  ")
	if r, err := RepoRoot(); err != nil || r != "/custom/root" {
		t.Fatalf("env root: %q %v", r, err)
	}
}

func TestReadBody_paths(t *testing.T) {
	root := testutil.SetRepoRoot(t)
	cat, err := Open(fixtureSkillsIndex)
	if err != nil {
		t.Fatal(err)
	}
	skillDir := filepath.Join(root, "pkg/playbook/testdata/skills/fixture-external")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	skillPath := filepath.Join(skillDir, "SKILL.md")
	body := "---\nname: x\n---\n\nBody after frontmatter.\n"
	if err := os.WriteFile(skillPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cat.byID["fixture-external"] = domain.SkillMeta{
		ID:           "fixture-external",
		ExternalPath: "pkg/playbook/testdata/skills/fixture-external/SKILL.md",
	}
	if _, err := cat.ReadBody("fixture-external"); err != nil {
		t.Fatal(err)
	}
	cat.byID["fixture-missing"] = domain.SkillMeta{
		ID:         "fixture-missing",
		CorpusPath: "pkg/playbook/testdata/skills/does-not-exist/SKILL.md",
	}
	if _, err := cat.ReadBody("fixture-missing"); err == nil {
		t.Fatal("expected read error")
	}
	huge := filepath.Join(root, "pkg/playbook/testdata/skills/fixture-huge")
	if err := os.MkdirAll(huge, 0o755); err != nil {
		t.Fatal(err)
	}
	hugeBody := strings.Repeat("x", MaxBodyBytes+100)
	if err := os.WriteFile(filepath.Join(huge, "SKILL.md"), []byte(hugeBody), 0o644); err != nil {
		t.Fatal(err)
	}
	cat.byID["fixture-huge"] = domain.SkillMeta{
		ID:         "fixture-huge",
		CorpusPath: "pkg/playbook/testdata/skills/fixture-huge/SKILL.md",
	}
	detail, err := cat.ReadBody("fixture-huge")
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Body) != MaxBodyBytes {
		t.Fatalf("truncated body: %d", len(detail.Body))
	}
}

func TestSearch_edgeCases(t *testing.T) {
	testutil.SetRepoRoot(t)
	cat, err := Open(fixtureSkillsIndex)
	if err != nil {
		t.Fatal(err)
	}
	all := cat.Search("", "", 0)
	if len(all) != 2 {
		t.Fatalf("empty query: %+v", all)
	}
	ranked := cat.Search("imaging disk forensic", "", 1)
	if len(ranked) != 1 || ranked[0].ID != "fixture-forensics" {
		t.Fatalf("ranking: %+v", ranked)
	}
	if len(cat.Search("no-match-xyz", "", 5)) != 0 {
		t.Fatal("expected no hits for nonsense query")
	}
}

func TestMappingsDir_repoError(t *testing.T) {
	defer SetRepoGetwd(func() (string, error) { return "", os.ErrInvalid })()
	t.Setenv(EnvRepoRoot, "")
	if _, err := MappingsDir(); err == nil {
		t.Fatal("expected error")
	}
}

func TestSkillMarkdownRel_externalPath(t *testing.T) {
	if got := skillMarkdownRel(domain.SkillMeta{CorpusPath: "a/b.md"}); got != "a/b.md" {
		t.Fatalf("corpus: %q", got)
	}
	if got := skillMarkdownRel(domain.SkillMeta{ExternalPath: "legacy/path.md"}); got != "legacy/path.md" {
		t.Fatalf("external: %q", got)
	}
}
