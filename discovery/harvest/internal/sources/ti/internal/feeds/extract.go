package feeds

import (
	"net"
	"net/url"
	"regexp"
	"strings"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
)

var (
	reMD5    = regexp.MustCompile(`(?i)\b[a-f0-9]{32}\b`)
	reSHA1   = regexp.MustCompile(`(?i)\b[a-f0-9]{40}\b`)
	reSHA256 = regexp.MustCompile(`(?i)\b[a-f0-9]{64}\b`)
	reURL    = regexp.MustCompile(`https?://[^\s"'<>]+`)
	reIPv4   = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)
)

// ExtractIOCsFromText pulls IOC-shaped tokens from unstructured text.
func ExtractIOCsFromText(s string) []domain.IOC {
	seen := map[string]struct{}{}
	var out []domain.IOC
	add := func(t domain.IOC) {
		k := string(t.Type) + ":" + t.Value
		if _, ok := seen[k]; ok {
			return
		}
		seen[k] = struct{}{}
		out = append(out, t)
	}

	for _, m := range reSHA256.FindAllString(s, -1) {
		add(domain.IOC{Type: domain.IOCHash, Value: strings.ToLower(m), Source: "extract"})
	}
	for _, m := range reSHA1.FindAllString(s, -1) {
		add(domain.IOC{Type: domain.IOCHash, Value: strings.ToLower(m), Source: "extract"})
	}
	for _, m := range reMD5.FindAllString(s, -1) {
		// avoid matching sha256 prefixes by only taking if not part of longer hex
		add(domain.IOC{Type: domain.IOCHash, Value: strings.ToLower(m), Source: "extract"})
	}
	for _, m := range reURL.FindAllString(s, -1) {
		if u, err := url.Parse(strings.TrimRight(m, ".,);")); err == nil && u.Host != "" {
			add(domain.IOC{Type: domain.IOCURL, Value: u.String(), Source: "extract"})
		}
	}
	for _, m := range reIPv4.FindAllString(s, -1) {
		if net.ParseIP(m) != nil {
			add(domain.IOC{Type: domain.IOCIP, Value: m, Source: "extract"})
		}
	}
	return out
}
