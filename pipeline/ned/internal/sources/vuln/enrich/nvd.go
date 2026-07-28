package enrich

import (
	"strings"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pipeline/pkg/nvd/parse"
	"github.com/butbeautifulv/veil/pkg/vuln/domain"
)

// FromNVDPage parses an NVD API page JSON and returns enriched vuln upsert envelopes.
func FromNVDPage(raw string) ([]*commit.Envelope, error) {
	parsed, _, err := parse.ParsePage([]byte(raw))
	if err != nil {
		return nil, err
	}
	var out []*commit.Envelope
	for _, p := range parsed {
		if strings.TrimSpace(p.CVE) == "" {
			continue
		}
		v := nvdToDomain(p)
		e, err := commit.NewEnvelope(commit.SourceVuln, commit.KindVulnUpsert, commit.VulnUpsertIdempotencyKey(v.CVE), v)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func nvdToDomain(p parse.Vulnerability) domain.Vulnerability {
	return domain.Vulnerability(p)
}
