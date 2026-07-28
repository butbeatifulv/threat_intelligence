package config

import (
	"testing"
)

func TestLoad_defaults(t *testing.T) {
	t.Setenv("VEIL_GRAPH_API_URL", "")
	t.Setenv("VEIL_ENGAGE_API_URL", "")
	t.Setenv("VEIL_GRAPH_MCP_URL", "")
	t.Setenv("VEIL_ENGAGE_MCP_URL", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GraphAPIURL.Host != "127.0.0.1:8090" {
		t.Fatalf("graph api host=%q", cfg.GraphAPIURL.Host)
	}
	if cfg.EngageAPIURL.Host != "127.0.0.1:8890" {
		t.Fatalf("engage api host=%q", cfg.EngageAPIURL.Host)
	}
	if cfg.GraphMCPURL.Host != "127.0.0.1:8091" {
		t.Fatalf("graph mcp host=%q", cfg.GraphMCPURL.Host)
	}
	if cfg.EngageMCPURL.Host != "127.0.0.1:8892" {
		t.Fatalf("engage mcp host=%q", cfg.EngageMCPURL.Host)
	}
}

func TestLoad_invalidURL(t *testing.T) {
	t.Setenv("VEIL_GRAPH_API_URL", "not-a-url")
	if _, err := Load(); err == nil {
		t.Fatal("expected error")
	}
}
