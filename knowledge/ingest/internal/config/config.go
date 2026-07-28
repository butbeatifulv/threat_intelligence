package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	NATSURL    string
	Stream     string
	Durable    string
	Subject    string
	Batch      int
	MaxWait    time.Duration
	Neo4jURI   string
	Neo4jUser  string
	Neo4jPass  string
	Neo4jDB    string
}

func Load() Config {
	return Config{
		NATSURL:   getenv("NATS_URL", "nats://localhost:4222"),
		Stream:    getenv("NATS_INGEST_STREAM", "INGEST"),
		Durable:   getenv("NATS_DURABLE", "ingest-worker"),
		Subject:   getenv("NATS_SUBSCRIBE_SUBJECT", "ingest.>"),
		Batch:     getenvInt("INGEST_BATCH", 10),
		MaxWait:   getenvDuration("INGEST_MAX_WAIT", 5*time.Second),
		Neo4jURI:  getenv("NEO4J_URI", "neo4j://localhost:7687"),
		Neo4jUser: getenv("NEO4J_USER", "neo4j"),
		Neo4jPass: getenv("NEO4J_PASS", "neo4jpassword"),
		Neo4jDB:   getenv("NEO4J_DB", "neo4j"),
	}
}

func getenv(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func getenvInt(k string, def int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getenvDuration(k string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
