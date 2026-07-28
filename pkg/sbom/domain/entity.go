// Package domain holds shared SBOM/advisory entities for scrape, NED, and graph ingest.
package domain

// AdvisoryRef points at an OSV/GHSA document for a CVE.
type AdvisoryRef struct {
	CVE    string
	Source string
	Path   string
}
