package components

import (
	"context"
	"fmt"
	"log/slog"

	coderulesneo "github.com/butbeautifulv/veil/knowledge/ingest/internal/appsec/coderules"
	nucleineo "github.com/butbeautifulv/veil/knowledge/ingest/internal/appsec/nuclei"
	sbomneo "github.com/butbeautifulv/veil/knowledge/ingest/internal/appsec/sbom"
	"github.com/butbeautifulv/veil/knowledge/ingest/internal/config"
	dsingest "github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/ds/envelope"
	engageingest "github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/engage/envelope"
	lolaingest "github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/lola/envelope"
	tiingest "github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/ti/envelope"
	vulningest "github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/vuln/envelope"
	"github.com/butbeautifulv/veil/pkg/commit"
)

// Runtime holds Neo4j writers and domain appliers for the ingest worker.
type Runtime struct {
	SBOM   *sbomneo.Store
	CR     *coderulesneo.Store
	Nuclei *nucleineo.Store
	Apply  DomainAppliers
	close  []func(context.Context) error
}

type DomainAppliers struct {
	TI     func(context.Context, *commit.Envelope) error
	Vuln   func(context.Context, *commit.Envelope) error
	Lola   func(context.Context, *commit.Envelope) error
	DS     func(context.Context, *commit.Envelope) error
	Engage func(context.Context, *commit.Envelope) error
}

func Init(ctx context.Context, cfg config.Config, log *slog.Logger) (*Runtime, error) {
	neoCfg := sbomneo.Config{
		URI: cfg.Neo4jURI, Username: cfg.Neo4jUser, Password: cfg.Neo4jPass, Database: cfg.Neo4jDB,
	}
	sbomSt, err := sbomneo.New(ctx, neoCfg)
	if err != nil {
		return nil, fmt.Errorf("neo4j sbom: %w", err)
	}
	crSt, err := coderulesneo.New(ctx, coderulesneo.Config(neoCfg))
	if err != nil {
		_ = sbomSt.Close(ctx)
		return nil, fmt.Errorf("neo4j coderules: %w", err)
	}
	nuSt, err := nucleineo.New(ctx, nucleineo.Config(neoCfg))
	if err != nil {
		_ = sbomSt.Close(ctx)
		_ = crSt.Close(ctx)
		return nil, fmt.Errorf("neo4j nuclei: %w", err)
	}

	wcfg := tiingest.NeoConfig{
		URI: neoCfg.URI, Username: neoCfg.Username, Password: neoCfg.Password, Database: neoCfg.Database,
	}
	tiEnsure, tiApply, tiClose, err := tiingest.SetupWriter(ctx, wcfg, log)
	if err != nil {
		_ = sbomSt.Close(ctx)
		_ = crSt.Close(ctx)
		_ = nuSt.Close(ctx)
		return nil, err
	}
	vulnEnsure, vulnApply, vulnClose, err := vulningest.SetupWriter(ctx, vulningest.NeoConfig{
		URI: neoCfg.URI, Username: neoCfg.Username, Password: neoCfg.Password, Database: neoCfg.Database,
	})
	if err != nil {
		_ = sbomSt.Close(ctx)
		_ = crSt.Close(ctx)
		_ = nuSt.Close(ctx)
		_ = tiClose(ctx)
		return nil, err
	}
	lolaEnsure, lolaApply, lolaClose, err := lolaingest.SetupWriter(ctx, lolaingest.NeoConfig{
		URI: neoCfg.URI, Username: neoCfg.Username, Password: neoCfg.Password, Database: neoCfg.Database,
	})
	if err != nil {
		_ = sbomSt.Close(ctx)
		_ = crSt.Close(ctx)
		_ = nuSt.Close(ctx)
		_ = tiClose(ctx)
		_ = vulnClose(ctx)
		return nil, err
	}
	dsEnsure, dsApply, dsClose, err := dsingest.SetupWriter(ctx, dsingest.NeoConfig{
		URI: neoCfg.URI, Username: neoCfg.Username, Password: neoCfg.Password, Database: neoCfg.Database,
	})
	if err != nil {
		_ = sbomSt.Close(ctx)
		_ = crSt.Close(ctx)
		_ = nuSt.Close(ctx)
		_ = tiClose(ctx)
		_ = vulnClose(ctx)
		_ = lolaClose(ctx)
		return nil, err
	}
	engageEnsure, engageApply, engageClose, err := engageingest.SetupWriter(ctx, engageingest.NeoConfig{
		URI: neoCfg.URI, Username: neoCfg.Username, Password: neoCfg.Password, Database: neoCfg.Database,
	})
	if err != nil {
		_ = sbomSt.Close(ctx)
		_ = crSt.Close(ctx)
		_ = nuSt.Close(ctx)
		_ = tiClose(ctx)
		_ = vulnClose(ctx)
		_ = lolaClose(ctx)
		_ = dsClose(ctx)
		return nil, err
	}

	for _, fn := range []func(context.Context) error{
		sbomSt.EnsureSchema, crSt.EnsureSchema, nuSt.EnsureSchema,
		tiEnsure, vulnEnsure, lolaEnsure, dsEnsure, engageEnsure,
	} {
		if err := fn(ctx); err != nil {
			rt := &Runtime{SBOM: sbomSt, CR: crSt, Nuclei: nuSt}
			rt.Shutdown(ctx)
			return nil, fmt.Errorf("schema: %w", err)
		}
	}

	return &Runtime{
		SBOM:   sbomSt,
		CR:     crSt,
		Nuclei: nuSt,
		Apply: DomainAppliers{
			TI: tiApply, Vuln: vulnApply, Lola: lolaApply, DS: dsApply, Engage: engageApply,
		},
		close: []func(context.Context) error{tiClose, vulnClose, lolaClose, dsClose, engageClose},
	}, nil
}

func (r *Runtime) Shutdown(ctx context.Context) {
	for _, fn := range r.close {
		_ = fn(ctx)
	}
	if r.Nuclei != nil {
		_ = r.Nuclei.Close(ctx)
	}
	if r.CR != nil {
		_ = r.CR.Close(ctx)
	}
	if r.SBOM != nil {
		_ = r.SBOM.Close(ctx)
	}
}
