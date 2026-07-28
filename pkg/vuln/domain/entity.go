// Package domain holds shared vulnerability entities for scrape, NED, and graph ingest.
package domain

type Vulnerability struct {
	ID       string       `json:"ID,omitempty"`
	CVE      string       `json:"CVE,omitempty"`
	Summary  string       `json:"Summary,omitempty"`
	CWE      []string     `json:"CWE,omitempty"`
	CPEs     []CPE        `json:"CPEs,omitempty"`
	CVSS     *CVSS        `json:"CVSS,omitempty"`
	Exploits []ExploitRef `json:"Exploits,omitempty"`
}

type CVSS struct {
	Version string  `json:"Version,omitempty"`
	Base    float64 `json:"Base,omitempty"`
	Vector  string  `json:"Vector,omitempty"`
}

type CPE struct {
	URI string `json:"URI,omitempty"`
}

type ExploitRef struct {
	Source string `json:"Source,omitempty"`
	RefID  string `json:"RefID,omitempty"`
	URL    string `json:"URL,omitempty"`
}
