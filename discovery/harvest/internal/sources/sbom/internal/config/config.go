package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds sbom scraper settings (env + defaults).
type Config struct {
	Sources     []string
	MaxCVE      int
	MaxGHSA     int
	GHSAMinYear int
	NATSURL       string
	ScrapeSubject string
	// CVEListFile / CVEListURL: OSV CVE ids (file wins if both set).
	CVEListFile string
	CVEListURL  string
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
	src := getenv("SBOM_SOURCES", "osv,ghsa")
	var parts []string
	for _, s := range strings.Split(src, ",") {
		s = strings.TrimSpace(strings.ToLower(s))
		if s != "" {
			parts = append(parts, s)
		}
	}
	if len(parts) == 0 {
		parts = []string{"osv", "ghsa"}
	}
	return &Config{
		Sources:     parts,
		MaxCVE:      getenvInt("SBOM_MAX_CVES", 200),
		MaxGHSA:     getenvInt("SBOM_MAX_GHSA", 100),
		GHSAMinYear: getenvInt("SBOM_GHSA_MIN_YEAR", 2017),
		NATSURL:     getenv("NATS_URL", "nats://localhost:4222"),
		ScrapeSubject: getenv("SBOM_SCRAPE_SUBJECT", "scrape.appsec.sbom"),
		CVEListFile: getenv("SBOM_CVE_LIST_FILE", ""),
		CVEListURL:  getenv("SBOM_CVE_LIST_URL", ""),
	}
}
