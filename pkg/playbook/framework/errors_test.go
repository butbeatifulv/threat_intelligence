package framework

import (
	"os"
	"path/filepath"
	"testing"

	pbindex "github.com/butbeautifulv/veil/pkg/playbook/index"
)

func TestLoadNavigatorLayer_errors(t *testing.T) {
	old := frameworkMappingsDir
	defer func() { frameworkMappingsDir = old }()
	frameworkMappingsDir = func() (string, error) { return "", os.ErrInvalid }
	if _, err := LoadNavigatorLayer(); err == nil {
		t.Fatal("expected mappings dir error")
	}

	root := t.TempDir()
	dir := filepath.Join(root, "pkg/playbook/corpus/mappings")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	frameworkMappingsDir = func() (string, error) { return dir, nil }
	if _, err := LoadNavigatorLayer(); err == nil {
		t.Fatal("expected read error")
	}
	if err := os.WriteFile(filepath.Join(dir, "attack-navigator-layer.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadNavigatorLayer(); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestLoadSubdomains_defaultsAndRepoError(t *testing.T) {
	old := frameworkRepoRoot
	defer func() { frameworkRepoRoot = old }()
	frameworkRepoRoot = func() (string, error) { return "", os.ErrInvalid }
	if _, err := LoadSubdomains(); err == nil {
		t.Fatal("expected repo error")
	}

	root := t.TempDir()
	idxDir := filepath.Join(root, filepath.Dir(pbindex.DefaultIndexRel))
	if err := os.MkdirAll(idxDir, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := `{"subdomain_counts":{"unknown-xyz":1}}`
	if err := os.WriteFile(filepath.Join(root, pbindex.DefaultIndexRel), []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	frameworkRepoRoot = func() (string, error) { return root, nil }
	subs, err := LoadSubdomains()
	if err != nil {
		t.Fatal(err)
	}
	if len(subs) != 1 || subs[0].Priority != "P3" || len(subs[0].VeilCats) != 1 || subs[0].VeilCats[0] != "playbook" {
		t.Fatalf("subs %#v", subs)
	}
}

func TestLoadSubdomains_readError(t *testing.T) {
	root := t.TempDir()
	old := frameworkRepoRoot
	defer func() { frameworkRepoRoot = old }()
	frameworkRepoRoot = func() (string, error) { return root, nil }
	if _, err := LoadSubdomains(); err == nil {
		t.Fatal("expected read error")
	}
}

func TestLoadSubdomains_sortByID(t *testing.T) {
	root := t.TempDir()
	idxDir := filepath.Join(root, filepath.Dir(pbindex.DefaultIndexRel))
	if err := os.MkdirAll(idxDir, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := `{"subdomain_counts":{"bbb":2,"aaa":2}}`
	if err := os.WriteFile(filepath.Join(root, pbindex.DefaultIndexRel), []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	old := frameworkRepoRoot
	defer func() { frameworkRepoRoot = old }()
	frameworkRepoRoot = func() (string, error) { return root, nil }
	subs, err := LoadSubdomains()
	if err != nil {
		t.Fatal(err)
	}
	if len(subs) != 2 || subs[0].ID != "aaa" || subs[1].ID != "bbb" {
		t.Fatalf("order %#v", subs)
	}
}

func TestLoadSubdomains_badJSON(t *testing.T) {
	root := t.TempDir()
	t.Setenv("VEIL_REPO_ROOT", root)
	dir := filepath.Join(root, "docs/skills-index")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cyber-skills.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSubdomains(); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestSkillsForTechnique_missingFile(t *testing.T) {
	root := t.TempDir()
	t.Setenv("VEIL_REPO_ROOT", root)
	if _, err := SkillsForTechnique("T1059"); err == nil {
		t.Fatal("expected read error")
	}
}

func TestSkillsForTechnique_repoRootError(t *testing.T) {
	old := frameworkRepoRoot
	defer func() { frameworkRepoRoot = old }()
	frameworkRepoRoot = func() (string, error) { return "", os.ErrInvalid }
	if _, err := SkillsForTechnique("T1059"); err == nil {
		t.Fatal("expected repo error")
	}
}

func TestSkillsForTechnique_badJSONAndNoMatch(t *testing.T) {
	root := t.TempDir()
	idxDir := filepath.Join(root, filepath.Dir(pbindex.DefaultIndexRel))
	if err := os.MkdirAll(idxDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, pbindex.DefaultIndexRel)
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := frameworkRepoRoot
	defer func() { frameworkRepoRoot = old }()
	frameworkRepoRoot = func() (string, error) { return root, nil }
	if _, err := SkillsForTechnique("T1059"); err == nil {
		t.Fatal("expected decode error")
	}

	doc := `{"skills":[{"id":"s1","attack_ids":["T1003"]}]}`
	if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	ids, err := SkillsForTechnique("T9999")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("got %v", ids)
	}
}

func TestNavigatorLayer_emptyTechniqueID(t *testing.T) {
	layer := &NavigatorLayer{
		Techniques: []NavigatorTechnique{{TechniqueID: "", Score: 1}, {TechniqueID: "T1", Score: 2}},
	}
	scores := layer.TechniqueScores()
	if len(scores) != 1 || scores["T1"] != 2 {
		t.Fatalf("scores %v", scores)
	}
	sum := layer.Summarize()
	if sum.AttackVersion != "" {
		t.Fatalf("version %q", sum.AttackVersion)
	}
}
