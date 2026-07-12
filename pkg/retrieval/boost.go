package retrieval

import (
	"regexp"
	"strings"
)

var (
	cveRe   = regexp.MustCompile(`(?i)\bCVE-\d{4}-\d+\b`)
	mitreRe = regexp.MustCompile(`\bT\d{4}(?:\.\d{3})?\b`)
	cweRe   = regexp.MustCompile(`(?i)\bCWE-\d+\b`)
)

const exactBoost = 2.0

// ApplyExactBoost increases scores when query contains cyber identifiers found in hit text.
func ApplyExactBoost(query string, ranked []RankedHit) []RankedHit {
	if len(ranked) == 0 {
		return ranked
	}
	tokens := extractIdentifiers(query)
	if len(tokens) == 0 {
		return ranked
	}
	out := make([]RankedHit, len(ranked))
	copy(out, ranked)
	for i := range out {
		hay := strings.ToLower(out[i].Hit.Text + " " + out[i].SkillID)
		for _, tok := range tokens {
			if strings.Contains(hay, strings.ToLower(tok)) {
				out[i].Score += exactBoost
			}
		}
	}
	sortRanked(out)
	return out
}

func extractIdentifiers(q string) []string {
	var out []string
	for _, re := range []*regexp.Regexp{cveRe, mitreRe, cweRe} {
		out = append(out, re.FindAllString(q, -1)...)
	}
	return out
}
