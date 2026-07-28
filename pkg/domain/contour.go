package domain

// Contour partitions Veil domain packages into ingest, engage, and knowledge/playbook.
type Contour string

const (
	ContourIngest    Contour = "ingest"
	ContourEngage    Contour = "engage"
	ContourKnowledge Contour = "knowledge"
)

// Valid reports whether c is a known contour value.
func (c Contour) Valid() bool {
	switch c {
	case ContourIngest, ContourEngage, ContourKnowledge:
		return true
	default:
		return false
	}
}
