// Package nvdmap holds helpers for NVD-derived vulnerability records.
package nvdmap

import "github.com/butbeautifulv/veil/pkg/vuln/domain"

// Vulnerability is the canonical NVD-derived vulnerability shape shared across layers.
type Vulnerability = domain.Vulnerability

// CVSS is the canonical CVSS sub-record.
type CVSS = domain.CVSS

// CPE is the canonical CPE sub-record.
type CPE = domain.CPE
