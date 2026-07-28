// Package hostnorm normalizes target strings for EngageTarget.name graph lookups.
package hostnorm

import "strings"

// NormalizeHost strips scheme, path, and port; lowercases host (matches knowledge/connector/query).
func NormalizeHost(target string) string {
	t := strings.TrimSpace(target)
	t = strings.TrimPrefix(t, "https://")
	t = strings.TrimPrefix(t, "http://")
	if i := strings.Index(t, "/"); i >= 0 {
		t = t[:i]
	}
	if i := strings.Index(t, ":"); i >= 0 {
		t = t[:i]
	}
	return strings.ToLower(strings.TrimSpace(t))
}
