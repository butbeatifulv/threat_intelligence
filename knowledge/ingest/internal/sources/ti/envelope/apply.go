// Package envelope applies commit TI messages to Neo4j (graph write path).
package envelope

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/butbeautifulv/veil/pkg/commit"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
	"github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/ti/jsonl"
	"github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/ti/repository"
	"github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/ti/usecase"
)

// ApplyEnvelope applies a TI envelope after Validate + source checks.
func ApplyEnvelope(ctx context.Context, repo repository.GraphRepository, uc *usecase.Ingestor, env *commit.Envelope) error {
	switch env.Kind {
	case commit.KindTIIoC:
		var i domain.IOC
		if err := json.Unmarshal(env.Payload, &i); err != nil {
			return err
		}
		_, err := uc.UpsertIOC(ctx, env.IdempotencyKey, i)
		return err
	case commit.KindTIKEVVulnerability:
		var p commit.TIKEVVulnPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return repo.UpsertKEVVulnerability(ctx, p.CVEID, p.VendorProject, p.Product, p.ShortDesc, p.DateAdded)
	case commit.KindTIReport:
		var r domain.Report
		if err := json.Unmarshal(env.Payload, &r); err != nil {
			return err
		}
		id, err := commit.ReportNodeID(env.IdempotencyKey)
		if err != nil {
			return err
		}
		r.ID = id
		return uc.UpsertReport(ctx, r)
	case commit.KindTICampaign:
		var c domain.Campaign
		if err := json.Unmarshal(env.Payload, &c); err != nil {
			return err
		}
		return uc.UpsertCampaign(ctx, c)
	case commit.KindTICluster:
		var cl domain.Cluster
		if err := json.Unmarshal(env.Payload, &cl); err != nil {
			return err
		}
		return uc.UpsertCluster(ctx, cl)
	case commit.KindTIActor:
		var a domain.Actor
		if err := json.Unmarshal(env.Payload, &a); err != nil {
			return err
		}
		id, err := commit.ActorNodeID(env.IdempotencyKey)
		if err != nil {
			return err
		}
		a.ID = id
		return uc.UpsertActor(ctx, a)
	case commit.KindTILinkCampaignIOC:
		var p commit.TILinkCampaignIOCPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		iocID, err := commit.IOCLinkNodeID(env.IdempotencyKey)
		if err != nil {
			return err
		}
		return repo.LinkCampaignIOC(ctx, p.CampaignID, iocID)
	case commit.KindTILinkClusterCampaign:
		var p commit.TILinkClusterCampaignPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		return repo.LinkClusterCampaign(ctx, p.ClusterID, p.CampaignID)
	case commit.KindTILinkCampaignActor:
		var p commit.TILinkCampaignActorPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		actorID := commit.TILinkSuffix(env.IdempotencyKey)
		if actorID == "" {
			return fmt.Errorf("ti link campaign actor: empty actor id in key %q", env.IdempotencyKey)
		}
		return repo.LinkCampaignActor(ctx, p.CampaignID, actorID, p.ActorName)
	case commit.KindTILinkReportMentionsIOC:
		var p commit.TILinkReportMentionsIOCPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		iocID, err := commit.IOCLinkNodeID(env.IdempotencyKey)
		if err != nil {
			return err
		}
		return repo.LinkReportMentionsIOC(ctx, p.ReportID, iocID)
	case commit.KindTIJSONLRecord:
		var p commit.TIJSONLRecordPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			return err
		}
		var lineEnv jsonl.Envelope
		if err := json.Unmarshal(p.Line, &lineEnv); err != nil {
			return fmt.Errorf("ti jsonl line: %w", err)
		}
		return uc.IngestOne(ctx, lineEnv)
	default:
		return fmt.Errorf("ti graph ingest: unsupported kind %q", env.Kind)
	}
}
