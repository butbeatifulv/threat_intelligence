package cvesource

import (
	"fmt"
	"strings"
)

// New returns a Lister from file path (preferred) or HTTP URL. At least one must be non-empty.
func New(filePath, httpURL string) (Lister, error) {
	fp := strings.TrimSpace(filePath)
	u := strings.TrimSpace(httpURL)
	if fp != "" {
		return FromFile{Path: fp}, nil
	}
	if u != "" {
		return FromHTTP{URL: u}, nil
	}
	return nil, fmt.Errorf("cvesource: set SBOM_CVE_LIST_FILE or SBOM_CVE_LIST_URL")
}
