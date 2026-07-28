package neo4jstore

import (
	"regexp"
	"strings"
)

var cvePattern = regexp.MustCompile(`(?i)CVE-\d{4}-\d{4,}`)

func extractCVEs(parts ...string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, p := range parts {
		for _, m := range cvePattern.FindAllString(p, -1) {
			id := strings.ToUpper(m)
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}
