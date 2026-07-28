package testutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRepoRoot_callerFails(t *testing.T) {
	old := repoCaller
	defer func() { repoCaller = old }()
	repoCaller = func(int) (uintptr, string, int, bool) { return 0, "", 0, false }
	if _, err := resolveRepoRoot(0); !errors.Is(err, errRepoCaller) {
		t.Fatalf("got %v", err)
	}
}

func TestMustRepoRoot_errors(t *testing.T) {
	f := &fakeTB{}
	mustRepoRoot(f, "", errRepoCaller)
	if !f.fatal {
		t.Fatal("expected fatal")
	}
	s := &fakeTB{}
	mustRepoRoot(s, "", errRepoRoot)
	if !s.skip {
		t.Fatal("expected skip")
	}
}

type fakeTB struct {
	fatal bool
	skip  bool
}

func (f *fakeTB) Helper()              {}
func (f *fakeTB) Fatal(args ...any)    { f.fatal = true }
func (f *fakeTB) Skip(args ...any)     { f.skip = true }

func TestResolveRepoRoot_versionsMissing(t *testing.T) {
	old := repoRootFrom
	defer func() { repoRootFrom = old }()
	repoRootFrom = func(string) string { return t.TempDir() }
	if _, err := resolveRepoRoot(0); !errors.Is(err, errRepoRoot) {
		t.Fatalf("got %v", err)
	}
}

func TestRepoRoot_ok(t *testing.T) {
	root := RepoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "versions.env")); err != nil {
		t.Fatalf("versions.env: %v", err)
	}
}

func TestRepoRoot_skipMissingVersions(t *testing.T) {
	old := repoRootFrom
	defer func() { repoRootFrom = old }()
	repoRootFrom = func(string) string { return t.TempDir() }
	t.Run("skip", func(t *testing.T) { RepoRoot(t) })
}

func TestSetRepoRoot_setsEnv(t *testing.T) {
	root := SetRepoRoot(t)
	if os.Getenv("VEIL_REPO_ROOT") != root {
		t.Fatalf("env: %q", os.Getenv("VEIL_REPO_ROOT"))
	}
}
