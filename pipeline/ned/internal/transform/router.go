package transform

import (
	"context"
	"fmt"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/sources/appsec"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/sources/ds"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/sources/lola"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/sources/ti"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/sources/vuln"
)

// ScrapeToIngest applies NED (normalize, enrich, dedup keys) and returns commit envelopes.
func ScrapeToIngest(ctx context.Context, env *harvest.Envelope) ([]*commit.Envelope, error) {
	switch env.Source {
	case harvest.SourceDS:
		return ds.Transform(ctx, env)
	case harvest.SourceTI:
		return ti.Transform(ctx, env)
	case harvest.SourceVuln:
		return vuln.Transform(ctx, env)
	case harvest.SourceLola:
		return lola.Transform(ctx, env)
	case harvest.SourceSBOM:
		return appsec.TransformSBOM(ctx, env)
	case harvest.SourceCoderules:
		return appsec.TransformCoderules(ctx, env)
	case harvest.SourceNuclei:
		return appsec.TransformNuclei(ctx, env)
	default:
		return nil, fmt.Errorf("pipeline: unknown source %q", env.Source)
	}
}
