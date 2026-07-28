package config

import "os"

type Config struct {
	Env   string
	Neo4j Neo4jConfig
}

type Neo4jConfig struct {
	URI      string
	Username string
	Password string
	Database string
}

func LoadConfig() (*Config, error) {
	return &Config{
		Env: "local",
		Neo4j: Neo4jConfig{
			URI:      envOr("NEO4J_URI", "neo4j://localhost:7687"),
			Username: envOr("NEO4J_USER", "neo4j"),
			Password: envOr("NEO4J_PASS", "neo4jpassword"),
			Database: envOr("NEO4J_DB", "neo4j"),
		},
	}, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
