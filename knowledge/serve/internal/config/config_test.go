package config

import (
	"os"
	"testing"
)

func TestLoadNeo4jURI(t *testing.T) {
	t.Setenv("NEO4J_USER", "neo4j")
	t.Setenv("NEO4J_PASS", "neo4jpassword")
	t.Setenv("NEO4J_DB", "neo4j")

	tests := []struct {
		name    string
		cluster string
		uri     string
		routing string
		want    string
	}{
		{
			name: "community default",
			want: "neo4j://localhost:7687",
		},
		{
			name:    "cluster default routing",
			cluster: "1",
			want:    "neo4j+routing://neo4j-core1:7687",
		},
		{
			name:    "cluster explicit uri",
			cluster: "1",
			uri:     "neo4j+routing://router.example:7687",
			want:    "neo4j+routing://router.example:7687",
		},
		{
			name:    "cluster routing override",
			cluster: "true",
			routing: "neo4j+routing://neo4j-core2:7687",
			want:    "neo4j+routing://neo4j-core2:7687",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("NEO4J_CLUSTER", tc.cluster)
			if tc.uri != "" {
				t.Setenv("NEO4J_URI", tc.uri)
			} else {
				t.Setenv("NEO4J_URI", "")
			}
			if tc.routing != "" {
				t.Setenv("NEO4J_ROUTING_URI", tc.routing)
			} else {
				t.Setenv("NEO4J_ROUTING_URI", "")
			}
			if got := loadNeo4jURI(); got != tc.want {
				t.Fatalf("loadNeo4jURI() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLoadAPIClusterNeo4j(t *testing.T) {
	t.Setenv("NEO4J_CLUSTER", "1")
	t.Setenv("NEO4J_URI", "")
	t.Setenv("API_LISTEN", ":8099")
	cfg := LoadAPI()
	if cfg.Neo4j.URI != "neo4j+routing://neo4j-core1:7687" {
		t.Fatalf("Neo4j.URI = %q", cfg.Neo4j.URI)
	}
	if cfg.ListenAddr != ":8099" {
		t.Fatalf("ListenAddr = %q", cfg.ListenAddr)
	}
}

func TestLoadMCPPreservesEnv(t *testing.T) {
	_ = os.Unsetenv("API_LISTEN")
	t.Setenv("NEO4J_CLUSTER", "0")
	t.Setenv("NEO4J_URI", "neo4j://neo4j:7687")
	cfg := LoadMCP()
	if cfg.Neo4j.URI != "neo4j://neo4j:7687" {
		t.Fatalf("Neo4j.URI = %q", cfg.Neo4j.URI)
	}
}
