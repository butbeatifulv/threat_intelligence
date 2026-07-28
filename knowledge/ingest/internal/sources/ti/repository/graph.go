package repository

import (
	"context"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
)

type GraphRepository interface {
	EnsureSchema(ctx context.Context) error

	UpsertIOC(ctx context.Context, id string, i domain.IOC) error
	UpsertCampaign(ctx context.Context, c domain.Campaign) error
	UpsertCluster(ctx context.Context, cl domain.Cluster) error
	UpsertActor(ctx context.Context, a domain.Actor) error
	UpsertReport(ctx context.Context, r domain.Report) error

	LinkCampaignIOC(ctx context.Context, campaignID, iocID string) error
	LinkClusterCampaign(ctx context.Context, clusterID string, campaignID string) error
	LinkCampaignActor(ctx context.Context, campaignID, actorID, actorName string) error
	LinkReportMentionsIOC(ctx context.Context, reportID, iocID string) error

	// KEV augments existing Vulnerability nodes (same label as vuln service).
	UpsertKEVVulnerability(ctx context.Context, cve, vendor, product, summary, dateAdded string) error
}
