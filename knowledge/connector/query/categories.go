package query

// Categories maps stable API/MCP category IDs to Neo4j labels used in this repo (scrapers).
var Categories = map[string]CategoryMeta{
	"vuln": {
		ID:          "vuln",
		Title:       "Vulnerabilities",
		Description: "CVEs, CWE, CPE, exploit metadata from vuln scraper",
		Labels:      []string{"Vulnerability", "CWE", "CPE", "Exploit"},
	},
	"ti": {
		ID:          "ti",
		Title:       "Threat intelligence",
		Description: "IOCs, actors, campaigns, clusters, reports from ti scraper",
		Labels:      []string{"IOC", "Campaign", "Cluster", "Actor", "Report"},
	},
	"detection": {
		ID:          "detection",
		Title:       "Detection content",
		Description: "Sigma, YARA, Atomic Red Team, Caldera abilities from ds scraper",
		Labels:      []string{"SigmaRule", "YaraRule", "AtomicTest", "CalderaAbility"},
	},
	"lola": {
		ID:          "lola",
		Title:       "LO(L)Bins & related",
		Description: "LOLBAS-style artifacts, commands, LOFTS entries from lola scraper",
		Labels:      []string{"LolaArtifact", "Command", "LoftsEntry"},
	},
	"mitre": {
		ID:          "mitre",
		Title:       "MITRE ATT&CK",
		Description: "ATT&CK tactics and techniques (STIX ingest)",
		Labels:      []string{"AttackTechnique", "AttackTactic"},
	},
	"sbom": {
		ID:          "sbom",
		Title:       "SBOM & advisories",
		Description: "Packages, GitHub security advisories, OSV links from sbom scraper",
		Labels:      []string{"Package", "SecurityAdvisory"},
	},
	"code_rules": {
		ID:          "code_rules",
		Title:       "Code rules & CWE catalog",
		Description: "Semgrep rules, CodeQL queries, CWE catalog enrichment from coderules scraper",
		Labels:      []string{"SemgrepRule", "CodeQLRule", "CWE"},
	},
	"dast": {
		ID:          "dast",
		Title:       "DAST / runtime templates",
		Description: "Nuclei templates (CVE-tagged HTTP checks) from nuclei scraper",
		Labels:      []string{"NucleiTemplate"},
	},
	"engage": {
		ID:          "engage",
		Title:       "Engage scans & findings",
		Description: "Tool runs and vulnerability findings from engage layer (cross-layer bus)",
		Labels:      []string{"EngageToolRun", "EngageFinding", "EngageTarget"},
	},
	"playbook": {
		ID:          "playbook",
		Title:       "Cybersecurity playbooks",
		Description: "Agent procedure skills (Anthropic corpus index); full text via playbook_get / GET /v1/playbooks/{id}",
		Labels:      []string{"CyberSkill"},
	},
}

// categoryOrder is the stable iteration order for APIs and docs.
var categoryOrder = []string{"vuln", "ti", "detection", "lola", "mitre", "sbom", "code_rules", "dast", "playbook", "engage"}

// CategoryIDs returns known category keys in stable order.
func CategoryIDs() []string {
	return append([]string(nil), categoryOrder...)
}

// ListCategoryMeta returns category descriptors in stable order.
func ListCategoryMeta() []CategoryMeta {
	out := make([]CategoryMeta, 0, len(categoryOrder))
	for _, id := range categoryOrder {
		if m, ok := Categories[id]; ok {
			out = append(out, m)
		}
	}
	return out
}

func labelsForCategory(category string) ([]string, bool) {
	c := trimCat(category)
	if c == "" {
		return nil, false
	}
	meta, ok := Categories[c]
	if !ok {
		return nil, false
	}
	return meta.Labels, true
}

func trimCat(s string) string {
	// minimal trim; callers validate alnum elsewhere
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

// ValidCategory reports whether id is a known category.
func ValidCategory(category string) bool {
	_, ok := Categories[trimCat(category)]
	return ok
}
