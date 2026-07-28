package appsec

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/sources/appsec/parse"

	"gopkg.in/yaml.v3"
)

// TransformSBOM maps harvest SBOM events to commit envelopes.
func TransformSBOM(ctx context.Context, env *harvest.Envelope) ([]*commit.Envelope, error) {
	_ = ctx
	switch env.Kind {
	case harvest.KindSBOMOSVJSON:
		var raw harvest.SBOMOSVRaw
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		var doc map[string]any
		if err := json.Unmarshal([]byte(raw.RawJSON), &doc); err != nil {
			return nil, err
		}
		affected, _ := doc["affected"].([]any)
		aff := make([]map[string]any, 0, len(affected))
		for _, a := range affected {
			if m, ok := a.(map[string]any); ok {
				aff = append(aff, m)
			}
		}
		pl := commit.SBOMOSVPayload{OSVID: raw.OSVID, CVE: raw.CVE, Affected: aff}
		key := commit.SBOMOSVIdempotencyKey(raw.CVE, "osv", raw.OSVID)
		out, err := commit.NewEnvelope(commit.SourceSBOM, commit.KindSBOMOSVRecord, key, pl)
		return []*commit.Envelope{out}, err

	case harvest.KindSBOMGHSAPath:
		var raw harvest.SBOMGHSARaw
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		pl := commit.SBOMGHSAPathPayload{Path: raw.Path, Doc: raw.Doc}
		out, err := commit.NewEnvelope(commit.SourceSBOM, commit.KindSBOMGHSADocument, commit.SBOMGHSAIdempotencyKey(raw.Path), pl)
		return []*commit.Envelope{out}, err

	default:
		return nil, fmt.Errorf("pipeline sbom: unknown kind %q", env.Kind)
	}
}

// TransformCoderules maps harvest coderules events to commit envelopes.
func TransformCoderules(ctx context.Context, env *harvest.Envelope) ([]*commit.Envelope, error) {
	_ = ctx
	switch env.Kind {
	case harvest.KindCoderulesCWERaw:
		var raw harvest.CoderulesCWERaw
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		pl := commit.CoderulesCWEPayload{ID: raw.ID, Name: raw.Name, Description: raw.Description, Status: raw.Status}
		out, err := commit.NewEnvelope(commit.SourceCoderules, commit.KindCoderulesCWERow, commit.CoderulesCWEIdempotencyKey(raw.ID), pl)
		return []*commit.Envelope{out}, err

	case harvest.KindCoderulesSemgrepRaw:
		return coderulesSemgrepRaw(env.Payload)

	case harvest.KindCoderulesCodeQLRaw:
		return coderulesCodeQLRaw(env.Payload)

	default:
		return nil, fmt.Errorf("pipeline coderules: unknown kind %q", env.Kind)
	}
}

// TransformNuclei maps harvest nuclei events to commit envelopes.
func TransformNuclei(ctx context.Context, env *harvest.Envelope) ([]*commit.Envelope, error) {
	_ = ctx
	if env.Kind != harvest.KindNucleiTemplateRaw {
		return nil, fmt.Errorf("pipeline nuclei: unknown kind %q", env.Kind)
	}
	return nucleiTemplateRaw(env.Payload)
}

func coderulesSemgrepRaw(payload json.RawMessage) ([]*commit.Envelope, error) {
	var raw harvest.CoderulesSemgrepRaw
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	var root map[string]any
	if err := yaml.Unmarshal([]byte(raw.RawYAML), &root); err != nil {
		return nil, err
	}
	ruleID, title := parse.SemgrepMeta(root, raw.Path)
	lang := strings.Split(raw.Path, "/")[0]
	cwes := parse.SemgrepCWES(root)
	pl := commit.CoderulesSemgrepPayload{
		Path: raw.Path, Language: lang, RuleID: ruleID, Title: title, RawYAML: raw.RawYAML, CWEs: cwes,
	}
	out, err := commit.NewEnvelope(commit.SourceCoderules, commit.KindCoderulesSemgrep, commit.CoderulesSemgrepIdempotencyKey(raw.Path), pl)
	return []*commit.Envelope{out}, err
}

func coderulesCodeQLRaw(payload json.RawMessage) ([]*commit.Envelope, error) {
	var raw harvest.CoderulesCodeQLRaw
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	cwes := parse.CodeQLCWES(raw.Body)
	pl := commit.CoderulesCodeQLPayload{
		Path: raw.Path, Name: raw.Path, Lang: "javascript", Body: raw.Body, CWEs: cwes,
	}
	out, err := commit.NewEnvelope(commit.SourceCoderules, commit.KindCoderulesCodeQL, commit.CoderulesCodeQLIdempotencyKey(raw.Path), pl)
	return []*commit.Envelope{out}, err
}

func nucleiTemplateRaw(payload json.RawMessage) ([]*commit.Envelope, error) {
	var raw harvest.NucleiTemplateRaw
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, err
	}
	p, err := parse.ParseNucleiYAML([]byte(raw.RawYAML))
	if err != nil || p.TemplateID == "" {
		return nil, err
	}
	pl := commit.NucleiTemplatePayload{
		Path: raw.Path, TemplateID: p.TemplateID, Name: p.Name, Severity: p.Severity,
		TagsJSON: p.TagsJSON, CVE: p.CVE, CWE: p.CWE, RawYAML: raw.RawYAML,
	}
	out, err := commit.NewEnvelope(commit.SourceNuclei, commit.KindNucleiTemplate, commit.NucleiTemplateIdempotencyKey(raw.Path), pl)
	return []*commit.Envelope{out}, err
}
