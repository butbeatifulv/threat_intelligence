package retrieval

import (
	"os"
	"path/filepath"

	pbindex "github.com/butbeautifulv/veil/pkg/playbook/index"
)

// VeilRoot resolves the Veil repository root for index paths.
func VeilRoot() (string, error) {
	if r := os.Getenv(EnvRepoRoot); r != "" {
		return r, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for i := 0; i < 14; i++ {
		if _, err := os.Stat(filepath.Join(dir, "versions.env")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "docs/skills-index/cyber-skills.json")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return pbindex.RepoRoot()
}
