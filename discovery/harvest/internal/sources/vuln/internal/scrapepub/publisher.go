package scrapepub

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/butbeautifulv/veil/pkg/harvest"
	"github.com/butbeautifulv/veil/pkg/vuln/domain"
	sharedpub "github.com/butbeautifulv/veil/discovery/harvest/internal/scrapepub"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/vuln/internal/repository"
	connats "github.com/butbeautifulv/veil/discovery/connector/nats"
)

type Publisher struct {
	sharedpub.Base
}

var _ repository.VulnerabilityRepository = (*Publisher)(nil)

func New(pub *connats.JetStreamPublisher, subject string) *Publisher {
	return NewFromRaw(sharedpub.NewRaw(pub, harvest.SourceVuln, subject))
}

func NewFromRaw(raw sharedpub.RawPublisher) *Publisher {
	return &Publisher{Base: sharedpub.NewBase(raw)}
}

func (p *Publisher) Save(ctx context.Context, v *domain.Vulnerability) error {
	return p.Upsert(ctx, v)
}

func (p *Publisher) FindByCVE(_ context.Context, _ string) (*domain.Vulnerability, error) {
	return nil, errors.New("vuln scrapepub: FindByCVE not supported")
}

func (p *Publisher) Upsert(ctx context.Context, v *domain.Vulnerability) error {
	if v == nil || strings.TrimSpace(v.CVE) == "" {
		return fmt.Errorf("vuln scrapepub: empty CVE")
	}
	key := "vuln:cve:" + strings.TrimSpace(strings.ToUpper(v.CVE))
	return p.Raw.Publish(ctx, harvest.KindVulnCVEUpsert, key, v)
}

func (p *Publisher) MergeExploitForCVE(ctx context.Context, cve string, ref domain.ExploitRef) error {
	cve = strings.TrimSpace(strings.ToUpper(cve))
	if cve == "" {
		return nil
	}
	pl := harvest.VulnMergeExploit{CVE: cve, Source: ref.Source, RefID: ref.RefID, URL: ref.URL}
	key := "vuln:exploit:" + cve + ":" + ref.Source + ":" + ref.RefID
	return p.Raw.Publish(ctx, harvest.KindVulnMergeExploit, key, pl)
}

func (p *Publisher) PublishNVDPage(ctx context.Context, startIndex int, rawJSON []byte) error {
	pl := harvest.VulnNVDPage{StartIndex: startIndex, RawJSON: string(rawJSON)}
	key := fmt.Sprintf("vuln:nvd:page:%d", startIndex)
	return p.Raw.Publish(ctx, harvest.KindVulnNVDPage, key, pl)
}
