package factory

import (
	"os"
	"strings"
)

func subjectForSource(name string) string {
	switch name {
	case "ds":
		return envOr("DS_SCRAPE_SUBJECT", "scrape.ds.events")
	case "vuln":
		return envOr("VULN_SCRAPE_SUBJECT", "scrape.vuln.events")
	case "lola":
		return envOr("LOLA_SCRAPE_SUBJECT", "scrape.lola.events")
	case "ti":
		return envOr("TI_SCRAPE_SUBJECT", "scrape.ti.events")
	case "sbom":
		return envOr("SBOM_SCRAPE_SUBJECT", "scrape.appsec.sbom")
	case "coderules":
		return envOr("CODERULES_SCRAPE_SUBJECT", "scrape.appsec.coderules")
	case "nuclei":
		return envOr("NUCLEI_SCRAPE_SUBJECT", "scrape.appsec.nuclei")
	default:
		return "scrape." + name + ".events"
	}
}

func scrapeSourceConstant(name string) (string, bool) {
	switch name {
	case "ds":
		return "ds", true
	case "vuln":
		return "vuln", true
	case "lola":
		return "lola", true
	case "ti":
		return "ti", true
	case "sbom":
		return "sbom", true
	case "coderules":
		return "coderules", true
	case "nuclei":
		return "nuclei", true
	default:
		return "", false
	}
}

func envOr(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return d
}
