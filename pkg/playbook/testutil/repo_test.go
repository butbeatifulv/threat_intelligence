package testutil

import (
	"os"
	"testing"
)

func TestSetRepoRoot(t *testing.T) {
	root := SetRepoRoot(t)
	if got := os.Getenv("VEIL_REPO_ROOT"); got != root {
		t.Fatalf("VEIL_REPO_ROOT=%q want %q", got, root)
	}
	if _, err := os.Stat(root + "/versions.env"); err != nil {
		t.Fatalf("versions.env: %v", err)
	}
}
