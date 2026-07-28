package usecase

import (
	"context"
	"log/slog"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
)

type fakeTIRepo struct {
	iocID    string
	ioc      domain.IOC
	campaign domain.Campaign
}

func (f *fakeTIRepo) EnsureSchema(ctx context.Context) error { return nil }
func (f *fakeTIRepo) UpsertIOC(ctx context.Context, id string, i domain.IOC) error {
	f.iocID, f.ioc = id, i
	return nil
}
func (f *fakeTIRepo) UpsertCampaign(ctx context.Context, c domain.Campaign) error {
	f.campaign = c
	return nil
}
func (f *fakeTIRepo) UpsertCluster(ctx context.Context, cl domain.Cluster) error { return nil }
func (f *fakeTIRepo) UpsertActor(ctx context.Context, a domain.Actor) error      { return nil }
func (f *fakeTIRepo) UpsertReport(ctx context.Context, r domain.Report) error    { return nil }
func (f *fakeTIRepo) LinkCampaignIOC(ctx context.Context, campaignID, iocID string) error {
	return nil
}
func (f *fakeTIRepo) LinkClusterCampaign(ctx context.Context, clusterID, campaignID string) error {
	return nil
}
func (f *fakeTIRepo) LinkCampaignActor(ctx context.Context, campaignID, actorID, actorName string) error {
	return nil
}
func (f *fakeTIRepo) LinkReportMentionsIOC(ctx context.Context, reportID, iocID string) error {
	return nil
}
func (f *fakeTIRepo) UpsertKEVVulnerability(ctx context.Context, cve, vendor, product, summary, dateAdded string) error {
	return nil
}

func TestUpsertIOC_emptyValue_skips(t *testing.T) {
	repo := &fakeTIRepo{}
	uc := NewIngestor(repo, slog.Default())
	ok, err := uc.UpsertIOC(context.Background(), commit.TIIoCIdempotencyKey("abc"), domain.IOC{Type: domain.IOCIP, Value: "  "})
	if err != nil {
		t.Fatal(err)
	}
	if ok || repo.iocID != "" {
		t.Fatalf("ok=%v iocID=%q", ok, repo.iocID)
	}
}

func TestUpsertIOC_usesIdempotencyNodeID(t *testing.T) {
	key := commit.TIIoCIdempotencyKey("deadbeef")
	wantID, err := commit.IOCNodeID(key)
	if err != nil {
		t.Fatal(err)
	}
	repo := &fakeTIRepo{}
	uc := NewIngestor(repo, slog.Default())
	ioc := domain.IOC{Type: domain.IOCDomain, Value: "evil.example"}
	ok, err := uc.UpsertIOC(context.Background(), key, ioc)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if repo.iocID != wantID {
		t.Fatalf("iocID %q want %q", repo.iocID, wantID)
	}
	if repo.ioc.Value != "evil.example" {
		t.Fatalf("ioc %+v", repo.ioc)
	}
}

func TestUpsertIOC_invalidKey(t *testing.T) {
	uc := NewIngestor(&fakeTIRepo{}, slog.Default())
	_, err := uc.UpsertIOC(context.Background(), "not-a-ti-key", domain.IOC{Type: domain.IOCIP, Value: "1.2.3.4"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpsertCampaign_requiresIDAndName(t *testing.T) {
	uc := NewIngestor(&fakeTIRepo{}, slog.Default())
	err := uc.UpsertCampaign(context.Background(), domain.Campaign{Name: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpsertCluster_requiresIDAndName(t *testing.T) {
	uc := NewIngestor(&fakeTIRepo{}, slog.Default())
	err := uc.UpsertCluster(context.Background(), domain.Cluster{Name: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpsertActor_requiresName(t *testing.T) {
	uc := NewIngestor(&fakeTIRepo{}, slog.Default())
	err := uc.UpsertActor(context.Background(), domain.Actor{ID: "a1"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpsertReport_requiresTitleAndLink(t *testing.T) {
	uc := NewIngestor(&fakeTIRepo{}, slog.Default())
	err := uc.UpsertReport(context.Background(), domain.Report{Title: "t"})
	if err == nil {
		t.Fatal("expected error")
	}
}
