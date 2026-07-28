package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	MaxTemplates         int
	YearsCSV             string
	NATSURL, ScrapeSubject string
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
	return &Config{
		MaxTemplates: getenvInt("NUCLEI_MAX", 120),
		YearsCSV:     getenv("NUCLEI_YEARS", "2023,2024,2025,2026"),
		NATSURL:      getenv("NATS_URL", "nats://localhost:4222"),
		ScrapeSubject: getenv("NUCLEI_SCRAPE_SUBJECT", "scrape.appsec.nuclei"),
	}
}
