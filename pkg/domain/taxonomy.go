package domain

// VeilCategory is a graph/read category label (playbook subdomain → category map).
type VeilCategory string

const (
	CategoryPlaybook   VeilCategory = "playbook"
	CategoryMITRE      VeilCategory = "mitre"
	CategoryDetection  VeilCategory = "detection"
	CategoryEngage     VeilCategory = "engage"
	CategoryVuln       VeilCategory = "vuln"
	CategoryTI         VeilCategory = "ti"
	CategoryAppSec     VeilCategory = "appsec"
	CategoryCompliance VeilCategory = "compliance"
)

// commonCategories lists frequently used VeilCategory values for validation helpers.
var commonCategories = []VeilCategory{
	CategoryPlaybook,
	CategoryMITRE,
	CategoryDetection,
	CategoryEngage,
	CategoryVuln,
	CategoryTI,
	CategoryAppSec,
	CategoryCompliance,
}

// CommonCategories returns the standard Veil category vocabulary.
func CommonCategories() []VeilCategory {
	out := make([]VeilCategory, len(commonCategories))
	copy(out, commonCategories)
	return out
}

// Valid reports whether c is a known common category (not an exhaustive graph enum).
func (c VeilCategory) Valid() bool {
	for _, known := range commonCategories {
		if c == known {
			return true
		}
	}
	return false
}

// String returns the wire/category string.
func (c VeilCategory) String() string { return string(c) }

// SubdomainFamily maps Anthropic skill subdomain ids to Veil categories (aligns playbook/framework).
var SubdomainFamily = map[string][]VeilCategory{
	"digital-forensics":          {CategoryPlaybook, CategoryMITRE},
	"incident-response":        {CategoryPlaybook, CategoryMITRE},
	"threat-hunting":             {CategoryPlaybook, CategoryDetection, CategoryMITRE},
	"malware-analysis":           {CategoryPlaybook, CategoryMITRE},
	"penetration-testing":        {CategoryPlaybook, CategoryEngage},
	"web-application-security":   {CategoryPlaybook, CategoryEngage, CategoryMITRE},
	"vulnerability-management":   {CategoryPlaybook, CategoryVuln},
	"threat-intelligence":        {CategoryPlaybook, CategoryTI},
	"cloud-security":             {CategoryPlaybook, CategoryEngage},
	"detection-engineering":      {CategoryPlaybook, CategoryDetection},
	"soc-operations":             {CategoryPlaybook, CategoryEngage},
}
