package retrieval

import "testing"

func TestConfigFromEnv_defaults(t *testing.T) {
	t.Setenv(EnvSearchEngine, "")
	t.Setenv(EnvSearchMode, "")
	cfg := ConfigFromEnv("/tmp/veil")
	if cfg.SearchEngine != "bleve" {
		t.Fatalf("engine: %q", cfg.SearchEngine)
	}
	if cfg.SearchMode != ModeHybrid {
		t.Fatalf("mode: %q", cfg.SearchMode)
	}
	if cfg.BleveIndexPath == "" {
		t.Fatal("expected bleve path")
	}
	if cfg.FetchK != DefaultFetchK {
		t.Fatalf("fetchK: %d", cfg.FetchK)
	}
}
