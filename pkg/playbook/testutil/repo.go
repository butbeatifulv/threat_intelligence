// Package testutil helps playbook tests locate the Veil repo root via VEIL_REPO_ROOT.
package testutil

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

var (
	repoCaller    = runtime.Caller
	repoRootFrom  = defaultRepoRootFrom
	errRepoCaller = errors.New("runtime.Caller failed")
	errRepoRoot   = errors.New("repo root not found (versions.env)")
)

type repoTB interface {
	Helper()
	Fatal(args ...any)
	Skip(args ...any)
}

func defaultRepoRootFrom(thisFile string) string {
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", ".."))
}

func resolveRepoRoot(skip int) (string, error) {
	_, file, _, ok := repoCaller(skip)
	if !ok {
		return "", errRepoCaller
	}
	root := repoRootFrom(file)
	if _, err := os.Stat(filepath.Join(root, "versions.env")); err != nil {
		return "", errRepoRoot
	}
	return root, nil
}

func mustRepoRoot(t repoTB, root string, err error) string {
	t.Helper()
	if err != nil {
		if errors.Is(err, errRepoCaller) {
			t.Fatal(err.Error())
		}
		t.Skip(err.Error())
	}
	return root
}

// RepoRoot returns the Veil monorepo root (directory containing versions.env).
func RepoRoot(t *testing.T) string {
	t.Helper()
	root, err := resolveRepoRoot(0)
	return mustRepoRoot(t, root, err)
}

// SetRepoRoot sets VEIL_REPO_ROOT for playbook loaders.
func SetRepoRoot(t *testing.T) string {
	t.Helper()
	root := RepoRoot(t)
	t.Setenv("VEIL_REPO_ROOT", root)
	return root
}
