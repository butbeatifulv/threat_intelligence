package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/pkg/harvest"
)

type Config struct {
	NATSURL         string
	ScrapeStream    string
	ScrapeDurable   string
	ScrapeSubject   string
	IngestPublish   string
	Batch           int
	MaxWait         time.Duration
	DomainSubjects  map[string]string
}

func Load() Config {
	return Config{
		NATSURL:       getenv("NATS_URL", "nats://localhost:4222"),
		ScrapeStream:  getenv("NATS_SCRAPE_STREAM", "SCRAPE"),
		ScrapeDurable: getenv("NATS_SCRAPE_DURABLE", "pipeline-worker"),
		ScrapeSubject: getenv("NATS_SCRAPE_SUBSCRIBE_SUBJECT", "scrape.>"),
		IngestPublish: getenv("NATS_INGEST_PUBLISH_SUBJECT", "ingest.events"),
		Batch:         getenvInt("PIPELINE_BATCH", 10),
		MaxWait:       getenvDuration("PIPELINE_MAX_WAIT", 5*time.Second),
		DomainSubjects: map[string]string{
			harvest.SourceDS:        strings.TrimSpace(os.Getenv("DS_INGEST_SUBJECT")),
			harvest.SourceTI:        strings.TrimSpace(os.Getenv("TI_INGEST_SUBJECT")),
			harvest.SourceVuln:      strings.TrimSpace(os.Getenv("VULN_INGEST_SUBJECT")),
			harvest.SourceLola:      strings.TrimSpace(os.Getenv("LOLA_INGEST_SUBJECT")),
			harvest.SourceSBOM:      strings.TrimSpace(os.Getenv("SBOM_INGEST_SUBJECT")),
			harvest.SourceCoderules: strings.TrimSpace(os.Getenv("CODERULES_INGEST_SUBJECT")),
			harvest.SourceNuclei:    strings.TrimSpace(os.Getenv("NUCLEI_INGEST_SUBJECT")),
		},
	}
}

func (c Config) IngestSubjectFor(source string) string {
	if s := c.DomainSubjects[source]; s != "" {
		return s
	}
	return c.IngestPublish
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
