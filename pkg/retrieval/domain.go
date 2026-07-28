package retrieval

// MatchType describes which retrieval path produced a hit.
type MatchType string

const (
	MatchKeyword MatchType = "keyword"
	MatchVector  MatchType = "vector"
	MatchHybrid  MatchType = "hybrid"
)

// SearchMode selects retrieval strategy.
type SearchMode string

const (
	ModeKeyword SearchMode = "keyword"
	ModeVector  SearchMode = "vector"
	ModeHybrid  SearchMode = "hybrid"
)

// ChunkHit is one indexed passage match.
type ChunkHit struct {
	SkillID      string
	Subdomain    string
	SectionTitle string
	ChunkIndex   int
	Text         string
	Score        float64
	MatchType    MatchType
}

// SearchResult is deduplicated skill-level hit.
type SearchResult struct {
	SkillID   string
	Score     float64
	Snippet   string
	MatchType MatchType
}

// SearchOpts query parameters.
type SearchOpts struct {
	Query     string
	Subdomain string
	Limit     int
	Mode      SearchMode
}

// RankedHit is a doc id with score for fusion.
type RankedHit struct {
	SkillID string
	Score   float64
	Hit     ChunkHit
}
