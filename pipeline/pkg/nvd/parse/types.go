package parse

import "github.com/butbeautifulv/veil/pkg/vuln/domain"

// Shared NVD wire types (canonical definitions live in pkg/vuln/domain).
type (
	Vulnerability = domain.Vulnerability
	CVSS          = domain.CVSS
	CPE           = domain.CPE
)
