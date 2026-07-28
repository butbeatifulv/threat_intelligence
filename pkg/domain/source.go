package domain

// Source is a stable ingest/commit source identifier (wire: harvest/commit envelope "source").
// Registry SOT: pkg/domain; harvest and commit consts mirror these strings.
type Source string

const (
	SourceSBOM      Source = "sbom"
	SourceCoderules Source = "coderules"
	SourceNuclei    Source = "nuclei"
	SourceTI        Source = "ti"
	SourceVuln      Source = "vuln"
	SourceLola      Source = "lola"
	SourceDS        Source = "ds"
	SourceBrowser   Source = "browser" // harvest-only scrape source
	SourceEngage    Source = "engage"   // commit-only graph ingest from engage events
)

// allSources is the canonical registry (union of harvest + commit well-known sources).
var allSources = []Source{
	SourceSBOM,
	SourceCoderules,
	SourceNuclei,
	SourceTI,
	SourceVuln,
	SourceLola,
	SourceDS,
	SourceBrowser,
	SourceEngage,
}

// AllSources returns every registered source id (stable order).
func AllSources() []Source {
	out := make([]Source, len(allSources))
	copy(out, allSources)
	return out
}

// Valid reports whether s is a registered source id.
func (s Source) Valid() bool {
	for _, known := range allSources {
		if s == known {
			return true
		}
	}
	return false
}

// String returns the wire string for s.
func (s Source) String() string { return string(s) }
