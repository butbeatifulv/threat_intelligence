package vuln

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
	"github.com/butbeautifulv/veil/pkg/vuln/domain"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/sources/vuln/enrich"
)

// Transform maps harvest vuln events to commit envelopes.
func Transform(ctx context.Context, env *harvest.Envelope) ([]*commit.Envelope, error) {
	_ = ctx
	switch env.Kind {
	case harvest.KindVulnCVEUpsert:
		var v domain.Vulnerability
		if err := json.Unmarshal(env.Payload, &v); err != nil {
			return nil, err
		}
		if strings.TrimSpace(v.CVE) == "" {
			return nil, fmt.Errorf("vuln: empty CVE")
		}
		out, err := commit.NewEnvelope(commit.SourceVuln, commit.KindVulnUpsert, commit.VulnUpsertIdempotencyKey(v.CVE), v)
		return []*commit.Envelope{out}, err

	case harvest.KindVulnMergeExploit:
		var raw harvest.VulnMergeExploit
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		p := commit.VulnMergeExploitPayload{CVE: raw.CVE, Source: raw.Source, RefID: raw.RefID, URL: raw.URL}
		key := commit.VulnMergeExploitIdempotencyKey(p.CVE, p.Source, p.RefID)
		out, err := commit.NewEnvelope(commit.SourceVuln, commit.KindVulnMergeExploit, key, p)
		return []*commit.Envelope{out}, err

	case harvest.KindVulnNVDPage:
		var page harvest.VulnNVDPage
		if err := json.Unmarshal(env.Payload, &page); err != nil {
			return nil, err
		}
		return enrich.FromNVDPage(page.RawJSON)

	default:
		return nil, fmt.Errorf("pipeline vuln: unknown kind %q", env.Kind)
	}
}
