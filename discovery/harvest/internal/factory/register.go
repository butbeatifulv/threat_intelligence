package factory

import (
	"fmt"
	"strings"
)

var sourceBuilders = map[string]func() Source{}

// Register adds a scrape source factory (call from source package init).
func Register(name string, fn func() Source) {
	sourceBuilders[strings.TrimSpace(strings.ToLower(name))] = fn
}

// SourcesFor returns sources for the given names (order preserved).
func SourcesFor(names []string) ([]Source, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("scrape: no sources requested")
	}
	out := make([]Source, 0, len(names))
	for _, raw := range names {
		name := strings.TrimSpace(strings.ToLower(raw))
		if name == "" {
			continue
		}
		newSrc, ok := sourceBuilders[name]
		if !ok {
			return nil, fmt.Errorf("unknown scrape source %q", name)
		}
		out = append(out, newSrc())
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("scrape: no valid sources in SCRAPE_SOURCES")
	}
	return out, nil
}

// ParseSourceNames splits SCRAPE_SOURCES (comma-separated).
func ParseSourceNames(env string) []string {
	if strings.TrimSpace(env) == "" {
		return []string{"ds"}
	}
	parts := strings.Split(env, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}
