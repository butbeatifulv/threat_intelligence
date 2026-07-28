package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/butbeautifulv/veil/knowledge/ingest/internal/components"
	coderulesneo "github.com/butbeautifulv/veil/knowledge/ingest/internal/appsec/coderules"
	nucleineo "github.com/butbeautifulv/veil/knowledge/ingest/internal/appsec/nuclei"
	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/natsjet"
	"github.com/butbeautifulv/veil/pkg/observability"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"
)

// RunPullLoop consumes JetStream messages until ctx is canceled.
func RunPullLoop(ctx context.Context, log *slog.Logger, sub *nats.Subscription, batch int, maxWait time.Duration, rt *components.Runtime) error {
	return natsjet.RunPullLoop(ctx, log, sub, natsjet.PullLoopOpts{
		Batch:    batch,
		MaxWait:  maxWait,
		NakDelay: 2 * time.Second,
		StopLog:  "consumer stopped",
	}, func(ctx context.Context, m *nats.Msg) error {
		return handleMsg(ctx, log, m, rt)
	})
}

func handleMsg(ctx context.Context, log *slog.Logger, m *nats.Msg, rt *components.Runtime) error {
	start := time.Now()
	status := "ok"
	subject := m.Subject
	ctx, span := observability.StartWorkerSpan(ctx, "ingest", "ingest.process_message",
		attribute.String("nats.subject", subject),
	)
	defer func() {
		observability.RecordNatsMessage("ingest", subject, status, time.Since(start).Seconds())
		span.End()
	}()

	var env commit.Envelope
	{
		decCtx, decSpan := observability.StartSpan(ctx, "ingest.decode_envelope")
		err := json.Unmarshal(m.Data, &env)
		decSpan.End()
		ctx = decCtx
		if err != nil {
			status = "error"
			observability.RecordError(span, err)
			return fmt.Errorf("decode envelope: %w", err)
		}
	}
	if err := env.Validate(); err != nil {
		status = "error"
		observability.RecordError(span, err)
		return err
	}
	if err := validateEnvelopeSource(&env); err != nil {
		status = "error"
		observability.RecordError(span, err)
		return err
	}
	span.SetAttributes(
		attribute.String("envelope.source", env.Source),
		attribute.String("envelope.kind", env.Kind),
	)

	applyCtx, applySpan := observability.StartSpan(ctx, "ingest.neo4j.apply."+env.Source)
	defer applySpan.End()

	var err error
	switch env.Source {
	case commit.SourceTI:
		err = rt.Apply.TI(applyCtx, &env)
	case commit.SourceVuln:
		err = rt.Apply.Vuln(applyCtx, &env)
	case commit.SourceLola:
		err = rt.Apply.Lola(applyCtx, &env)
	case commit.SourceDS:
		err = rt.Apply.DS(applyCtx, &env)
	case commit.SourceEngage:
		err = rt.Apply.Engage(applyCtx, &env)
	default:
		err = applyByKind(applyCtx, log, &env, rt)
	}
	if err != nil {
		status = "error"
		observability.RecordError(span, err)
		observability.RecordError(applySpan, err)
		return err
	}
	return nil
}

func applyByKind(ctx context.Context, log *slog.Logger, env *commit.Envelope, rt *components.Runtime) error {
	switch env.Kind {
	case commit.KindSBOMOSVRecord:
		var p commit.SBOMOSVPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return rt.SBOM.UpsertFromOSVVuln(ctx, p.OSVID, p.CVE, p.Affected)
	case commit.KindSBOMGHSADocument:
		var p commit.SBOMGHSAPathPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		if p.Doc == nil {
			return fmt.Errorf("ghsa: empty doc")
		}
		return rt.SBOM.UpsertGHSA(ctx, p.Doc)
	case commit.KindCoderulesCWERow:
		var p commit.CoderulesCWEPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return rt.CR.UpsertCWECatalog(ctx, p.ID, p.Name, p.Description, p.Status)
	case commit.KindCoderulesSemgrep:
		var p commit.CoderulesSemgrepPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		md := fmt.Sprintf("# %s\n\n**path:** `%s`\n\n```yaml\n%s\n```\n", p.Title, p.Path, p.RawYAML)
		id := coderulesneo.StableID("semgrep", p.Path)
		if err := rt.CR.UpsertSemgrepRule(ctx, id, p.Path, p.Title, p.Language, md); err != nil {
			return err
		}
		for _, cw := range p.CWEs {
			if err := rt.CR.LinkSemgrepRuleToCWE(ctx, id, cw); err != nil {
				return err
			}
		}
		return nil
	case commit.KindCoderulesCodeQL:
		var p commit.CoderulesCodeQLPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		md := fmt.Sprintf("# %s\n\n**path:** `%s`\n\n```ql\n%s\n```\n", p.Name, p.Path, p.Body)
		id := coderulesneo.StableID("codeql", p.Path)
		if err := rt.CR.UpsertCodeQLRule(ctx, id, p.Path, p.Name, p.Lang, md); err != nil {
			return err
		}
		for _, cw := range p.CWEs {
			if err := rt.CR.LinkCodeQLRuleToCWE(ctx, id, cw); err != nil {
				return err
			}
		}
		return nil
	case commit.KindNucleiTemplate:
		var p commit.NucleiTemplatePayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		id := nucleineo.StableID("nuclei", p.Path)
		md := fmt.Sprintf("# %s\n\n**id:** `%s`  \n**path:** `%s`\n\n```yaml\n%s\n```\n", p.Name, p.TemplateID, p.Path, p.RawYAML)
		return rt.Nuclei.UpsertNucleiTemplate(ctx, id, p.TemplateID, p.Path, p.Name, p.Severity, p.TagsJSON, p.CVE, p.CWE, md)
	default:
		log.Warn("unknown kind", slog.String("kind", env.Kind), slog.String("source", env.Source))
		return nil
	}
}

func validateEnvelopeSource(e *commit.Envelope) error {
	switch e.Kind {
	case commit.KindSBOMOSVRecord, commit.KindSBOMGHSADocument:
		if e.Source != commit.SourceSBOM {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, commit.SourceSBOM, e.Source)
		}
	case commit.KindCoderulesCWERow, commit.KindCoderulesSemgrep, commit.KindCoderulesCodeQL:
		if e.Source != commit.SourceCoderules {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, commit.SourceCoderules, e.Source)
		}
	case commit.KindNucleiTemplate:
		if e.Source != commit.SourceNuclei {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, commit.SourceNuclei, e.Source)
		}
	case commit.KindTIIoC, commit.KindTIKEVVulnerability, commit.KindTIReport, commit.KindTICampaign, commit.KindTICluster, commit.KindTIActor,
		commit.KindTILinkCampaignIOC, commit.KindTILinkClusterCampaign, commit.KindTILinkCampaignActor, commit.KindTILinkReportMentionsIOC, commit.KindTIJSONLRecord:
		if e.Source != commit.SourceTI {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, commit.SourceTI, e.Source)
		}
	case commit.KindVulnUpsert, commit.KindVulnMergeExploit:
		if e.Source != commit.SourceVuln {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, commit.SourceVuln, e.Source)
		}
	case commit.KindLolaArtifact, commit.KindLolaLofts, commit.KindLolaAttackTechnique, commit.KindLolaAttackTactic,
		commit.KindLolaMergeTacticTechnique, commit.KindLolaMergeSubtechnique, commit.KindLolaLinkArtifacts:
		if e.Source != commit.SourceLola {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, commit.SourceLola, e.Source)
		}
	case commit.KindDSUpsertSigma, commit.KindDSUpsertYara, commit.KindDSUpsertAtomic, commit.KindDSUpsertCaldera:
		if e.Source != commit.SourceDS {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, commit.SourceDS, e.Source)
		}
	case commit.KindEngageToolRun, commit.KindEngageFinding:
		if e.Source != commit.SourceEngage {
			return fmt.Errorf("kind %q expects source %q, got %q", e.Kind, commit.SourceEngage, e.Source)
		}
	}
	return nil
}
