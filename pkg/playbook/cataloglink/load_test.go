package cataloglink

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCatalog_emptyRepo(t *testing.T) {
	ResetCatalogForTest()
	t.Setenv("VEIL_REPO_ROOT", t.TempDir())
	catalogOnce.Do(loadCatalog)
	if len(catalogNames) != 0 {
		t.Fatalf("expected empty catalog, got %d", len(catalogNames))
	}
	ResetCatalogForTest()
}

func TestLoadCatalog_repoRootError(t *testing.T) {
	ResetCatalogForTest()
	old := catalogRepoRoot
	defer func() { catalogRepoRoot = old }()
	catalogRepoRoot = func() (string, error) { return "", os.ErrInvalid }
	catalogOnce.Do(loadCatalog)
	if catalogNames == nil || len(catalogNames) != 0 {
		t.Fatalf("catalogNames = %#v", catalogNames)
	}
	ResetCatalogForTest()
}

func TestLoadCatalog_parsesToolsYAML(t *testing.T) {
	root := t.TempDir()
	catDir := filepath.Join(root, "engage", "serve", "catalog")
	if err := os.MkdirAll(catDir, 0o755); err != nil {
		t.Fatal(err)
	}
	yaml := "tools:\n  - name: demo_tool\n    binary: demo_bin\n"
	if err := os.WriteFile(filepath.Join(catDir, "tools.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	ResetCatalogForTest()
	t.Setenv("VEIL_REPO_ROOT", root)
	catalogOnce.Do(loadCatalog)
	if _, ok := catalogNames["demo_tool"]; !ok {
		t.Fatal("missing demo_tool")
	}
	if _, ok := catalogNames["demo_bin"]; !ok {
		t.Fatal("missing demo_bin")
	}
	ResetCatalogForTest()
}
