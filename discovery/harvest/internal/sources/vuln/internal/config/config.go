package config

import "os"

type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	Neo4j      Neo4jConfig
	NVD        NVDConfig
}

type Neo4jConfig struct {
	URI      string
	Username string
	Password string
	Database string
}

type NVDConfig struct {
	APIKey string
}

func LoadConfig() (*Config, error) {
	// Minimal config loader for local usage.
	// Services use Neo4j only (graph) and can be configured via env.
	return &Config{
		Env: "local",
		Neo4j: Neo4jConfig{
			URI:      envOr("NEO4J_URI", "neo4j://localhost:7687"),
			Username: envOr("NEO4J_USER", "neo4j"),
			Password: envOr("NEO4J_PASS", "neo4jpassword"),
			Database: envOr("NEO4J_DB", "neo4j"),
		},
		NVD: NVDConfig{
			APIKey: envOr("NVD_API_KEY", ""),
		},
	}, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
