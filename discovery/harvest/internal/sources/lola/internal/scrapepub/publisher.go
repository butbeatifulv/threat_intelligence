package scrapepub

import (
	"context"
	"encoding/json"

	"github.com/butbeautifulv/veil/pkg/harvest"
	"github.com/butbeautifulv/veil/pkg/lola/domain"
	sharedpub "github.com/butbeautifulv/veil/discovery/harvest/internal/scrapepub"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/lola/internal/repository"
	connats "github.com/butbeautifulv/veil/discovery/connector/nats"
)

type Publisher struct {
	sharedpub.Base
}

var _ repository.LolaRepository = (*Publisher)(nil)

func New(pub *connats.JetStreamPublisher, subject string) *Publisher {
	return NewFromRaw(sharedpub.NewRaw(pub, harvest.SourceLola, subject))
}

func NewFromRaw(raw sharedpub.RawPublisher) *Publisher {
	return &Publisher{Base: sharedpub.NewBase(raw)}
}

func (p *Publisher) EnsureSchema(_ context.Context) error { return nil }

func (p *Publisher) UpsertArtifact(ctx context.Context, source string, a *domain.Artifact) error {
	if a == nil {
		return nil
	}
	body, err := json.Marshal(a)
	if err != nil {
		return err
	}
	pl := harvest.LolaArtifactRaw{Source: source, Path: a.Name, RawBody: string(body)}
	key := "lola:artifact:" + source + ":" + a.Name
	return p.Raw.Publish(ctx, harvest.KindLolaArtifactRaw, key, pl)
}

func (p *Publisher) UpsertLoftsEntry(ctx context.Context, title, category, linkURL, markdown string) error {
	pl := harvest.LolaLoftsRaw{Title: title, Category: category, LinkURL: linkURL, Markdown: markdown}
	return p.Raw.Publish(ctx, harvest.KindLolaLoftsRaw, "lola:lofts:"+linkURL, pl)
}

func (p *Publisher) UpsertAttackTechnique(ctx context.Context, id, name, description, markdown string) error {
	pl := harvest.LolaAttackTechnique{ID: id, Name: name, Description: description, Markdown: markdown}
	return p.Raw.Publish(ctx, harvest.KindLolaAttackTechnique, "lola:technique:"+id, pl)
}

func (p *Publisher) UpsertAttackTactic(ctx context.Context, id, name, description, markdown string) error {
	pl := harvest.LolaAttackTactic{ID: id, Name: name, Description: description, Markdown: markdown}
	return p.Raw.Publish(ctx, harvest.KindLolaAttackTactic, "lola:tactic:"+id, pl)
}

func (p *Publisher) MergeTacticIncludesTechnique(ctx context.Context, tacticID, techniqueID string) error {
	pl := harvest.LolaMergeTacticTechnique{TacticID: tacticID, TechniqueID: techniqueID}
	key := "lola:merge:" + tacticID + ":" + techniqueID
	return p.Raw.Publish(ctx, harvest.KindLolaMergeTacticTechnique, key, pl)
}

func (p *Publisher) MergeSubtechniqueOf(ctx context.Context, parentTechniqueID, childTechniqueID string) error {
	pl := harvest.LolaMergeSubtechnique{ParentTechniqueID: parentTechniqueID, ChildTechniqueID: childTechniqueID}
	key := "lola:sub:" + parentTechniqueID + ":" + childTechniqueID
	return p.Raw.Publish(ctx, harvest.KindLolaMergeSubtechnique, key, pl)
}

func (p *Publisher) LinkArtifactsAndCommandsToTechniques(ctx context.Context) error {
	return p.Raw.Publish(ctx, harvest.KindLolaLinkArtifacts, "lola:link-artifacts", struct{}{})
}
