// Package cache provides the default scrape blob directory on the host.
package cache

import (
	"os"
	"path/filepath"
	"strings"
)

// DefaultDir returns SCRAPE_CACHE_DIR or empty (compose sets /data/cache in containers).
func DefaultDir() string {
	if v := strings.TrimSpace(os.Getenv("SCRAPE_CACHE_DIR")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("VEIL_ROOT")); v != "" {
		return filepath.Join(v, "var", "veil", "blobs")
	}
	return filepath.Join(".", "var", "veil", "blobs")
}
