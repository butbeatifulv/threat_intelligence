package lola

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
)

// Transform maps harvest lola events to commit envelopes.
func Transform(ctx context.Context, env *harvest.Envelope) ([]*commit.Envelope, error) {
	_ = ctx
	switch env.Kind {
	case harvest.KindLolaArtifactRaw:
		var raw harvest.LolaArtifactRaw
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		pl := commit.LolaArtifactPayload{Source: raw.Source, Body: json.RawMessage(raw.RawBody)}
		key := commit.LolaArtifactIdempotencyKey(raw.Source, artifactNameFromBody(raw.RawBody))
		out, err := commit.NewEnvelope(commit.SourceLola, commit.KindLolaArtifact, key, pl)
		return []*commit.Envelope{out}, err

	case harvest.KindLolaLoftsRaw:
		var raw harvest.LolaLoftsRaw
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		pl := commit.LolaLoftsPayload{Title: raw.Title, Category: raw.Category, LinkURL: raw.LinkURL, Markdown: raw.Markdown}
		out, err := commit.NewEnvelope(commit.SourceLola, commit.KindLolaLofts, commit.LolaLoftsIdempotencyKey(raw.LinkURL), pl)
		return []*commit.Envelope{out}, err

	case harvest.KindLolaLinkArtifacts:
		pl := commit.LolaLinkArtifactsPayload{}
		out, err := commit.NewEnvelope(commit.SourceLola, commit.KindLolaLinkArtifacts, commit.LolaLinkArtifactsIdempotencyKey(), pl)
		return []*commit.Envelope{out}, err

	case harvest.KindLolaAttackTechnique:
		var raw harvest.LolaAttackTechnique
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		p := commit.LolaAttackTechniquePayload{ID: raw.ID, Name: raw.Name, Description: raw.Description, Markdown: raw.Markdown}
		out, err := commit.NewEnvelope(commit.SourceLola, commit.KindLolaAttackTechnique, commit.LolaTechniqueIdempotencyKey(p.ID), p)
		return []*commit.Envelope{out}, err

	case harvest.KindLolaAttackTactic:
		var raw harvest.LolaAttackTactic
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		p := commit.LolaAttackTacticPayload{ID: raw.ID, Name: raw.Name, Description: raw.Description, Markdown: raw.Markdown}
		out, err := commit.NewEnvelope(commit.SourceLola, commit.KindLolaAttackTactic, commit.LolaTacticIdempotencyKey(p.ID), p)
		return []*commit.Envelope{out}, err

	case harvest.KindLolaMergeTacticTechnique:
		var raw harvest.LolaMergeTacticTechnique
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		p := commit.LolaMergeTacticTechniquePayload{TacticID: raw.TacticID, TechniqueID: raw.TechniqueID}
		out, err := commit.NewEnvelope(commit.SourceLola, commit.KindLolaMergeTacticTechnique, commit.LolaMergeTacticTechniqueIdempotencyKey(p.TacticID, p.TechniqueID), p)
		return []*commit.Envelope{out}, err

	case harvest.KindLolaMergeSubtechnique:
		var raw harvest.LolaMergeSubtechnique
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		p := commit.LolaMergeSubtechniquePayload{ParentTechniqueID: raw.ParentTechniqueID, ChildTechniqueID: raw.ChildTechniqueID}
		out, err := commit.NewEnvelope(commit.SourceLola, commit.KindLolaMergeSubtechnique, commit.LolaMergeSubtechniqueIdempotencyKey(p.ParentTechniqueID, p.ChildTechniqueID), p)
		return []*commit.Envelope{out}, err

	default:
		return nil, fmt.Errorf("pipeline lola: unknown kind %q", env.Kind)
	}
}

func artifactNameFromBody(body string) string {
	var a struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal([]byte(body), &a)
	return a.Name
}
