package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Sources                       []string
	MaxCWE, MaxSemgrep, MaxCodeQL int
	NATSURL, ScrapeSubject        string
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

func FromEnv() *Config {
	src := getenv("CODERULES_SOURCES", "cwe,semgrep,codeql")
	var parts []string
	for _, s := range strings.Split(src, ",") {
		s = strings.TrimSpace(strings.ToLower(s))
		if s != "" {
			parts = append(parts, s)
		}
	}
	if len(parts) == 0 {
		parts = []string{"cwe", "semgrep", "codeql"}
	}
	return &Config{
		Sources:     parts,
		MaxCWE:      getenvInt("CODERULES_MAX_CWE", 5000),
		MaxSemgrep:  getenvInt("CODERULES_MAX_SEMGREP", 80),
		MaxCodeQL:   getenvInt("CODERULES_MAX_CODEQL", 60),
		NATSURL:     getenv("NATS_URL", "nats://localhost:4222"),
		ScrapeSubject: getenv("CODERULES_SCRAPE_SUBJECT", "scrape.appsec.coderules"),
	}
}
