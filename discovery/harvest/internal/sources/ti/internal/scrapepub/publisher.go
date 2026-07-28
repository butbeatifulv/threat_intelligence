// Package scrapepub publishes raw TI events to scrape.> for pipeline-worker.
package scrapepub

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/butbeautifulv/veil/pkg/harvest"
	"github.com/butbeautifulv/veil/pkg/ti/domain"
	sharedpub "github.com/butbeautifulv/veil/discovery/harvest/internal/scrapepub"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/ti/internal/repository"
	connats "github.com/butbeautifulv/veil/discovery/connector/nats"
)

type Publisher struct {
	sharedpub.Base
}

var _ repository.GraphRepository = (*Publisher)(nil)

func New(pub *connats.JetStreamPublisher, subject string) *Publisher {
	return NewFromRaw(sharedpub.NewRaw(pub, harvest.SourceTI, subject))
}

func NewFromRaw(raw sharedpub.RawPublisher) *Publisher {
	return &Publisher{Base: sharedpub.NewBase(raw)}
}

func (p *Publisher) EnsureSchema(_ context.Context) error { return nil }

func (p *Publisher) UpsertIOC(ctx context.Context, i domain.IOC) error {
	key := "ti:ioc:" + strings.TrimSpace(string(i.Type)) + ":" + strings.TrimSpace(i.Value)
	return p.Raw.Publish(ctx, harvest.KindTIIoCRaw, key, i)
}

func (p *Publisher) UpsertCampaign(ctx context.Context, c domain.Campaign) error {
	key := "ti:campaign:" + strings.TrimSpace(c.ID)
	return p.Raw.Publish(ctx, harvest.KindTICampaignRaw, key, c)
}

func (p *Publisher) UpsertCluster(ctx context.Context, cl domain.Cluster) error {
	key := "ti:cluster:" + strings.TrimSpace(cl.ID)
	return p.Raw.Publish(ctx, harvest.KindTIClusterRaw, key, cl)
}

func (p *Publisher) UpsertActor(ctx context.Context, a domain.Actor) error {
	key := "ti:actor:" + strings.TrimSpace(a.Name)
	return p.Raw.Publish(ctx, harvest.KindTIActorRaw, key, a)
}

func (p *Publisher) UpsertReport(ctx context.Context, r domain.Report) error {
	key := "ti:report:" + strings.TrimSpace(r.Link)
	return p.Raw.Publish(ctx, harvest.KindTIReportRaw, key, r)
}

func (p *Publisher) LinkCampaignIOC(ctx context.Context, campaignID string, i domain.IOC) error {
	_ = campaignID
	return p.UpsertIOC(ctx, i)
}

func (p *Publisher) LinkClusterCampaign(ctx context.Context, clusterID, campaignID string) error {
	_ = clusterID
	_ = campaignID
	return nil
}

func (p *Publisher) LinkCampaignActor(ctx context.Context, campaignID, actorName string) error {
	_ = campaignID
	return p.UpsertActor(ctx, domain.Actor{Name: actorName})
}

func (p *Publisher) LinkReportMentionsIOC(ctx context.Context, reportID string, i domain.IOC) error {
	_ = reportID
	return p.UpsertIOC(ctx, i)
}

func (p *Publisher) UpsertKEVVulnerability(ctx context.Context, cve, vendor, product, summary, dateAdded string) error {
	pl := harvest.TIKEVRow{
		CVEID: cve, VendorProject: vendor, Product: product, ShortDesc: summary, DateAdded: dateAdded,
	}
	cveU := strings.TrimSpace(strings.ToUpper(cve))
	return p.Raw.Publish(ctx, harvest.KindTIKEVRow, "ti:kev:"+cveU, pl)
}

// PublishJSONLLine publishes one raw JSONL line.
func (p *Publisher) PublishJSONLLine(ctx context.Context, line []byte) error {
	sum := sha256.Sum256(line)
	key := "ti:jsonl:" + hex.EncodeToString(sum[:])
	pl := harvest.TIJSONLLine{Line: json.RawMessage(line)}
	return p.Raw.Publish(ctx, harvest.KindTIJSONLLine, key, pl)
}
