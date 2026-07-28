package ti

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
	tidomain "github.com/butbeautifulv/veil/pkg/ti/domain"
	tinormalize "github.com/butbeautifulv/veil/pkg/ti/normalize"
)

// Transform maps harvest TI events to commit envelopes.
func Transform(ctx context.Context, env *harvest.Envelope) ([]*commit.Envelope, error) {
	_ = ctx
	switch env.Kind {
	case harvest.KindTIKEVRow:
		var row harvest.TIKEVRow
		if err := json.Unmarshal(env.Payload, &row); err != nil {
			return nil, err
		}
		p := commit.TIKEVVulnPayload{
			CVEID: row.CVEID, VendorProject: row.VendorProject, Product: row.Product,
			ShortDesc: row.ShortDesc, DateAdded: row.DateAdded,
		}
		cveU := strings.TrimSpace(strings.ToUpper(p.CVEID))
		out, err := commit.NewEnvelope(commit.SourceTI, commit.KindTIKEVVulnerability, commit.TIKEVIdempotencyKey(cveU), p)
		return []*commit.Envelope{out}, err

	case harvest.KindTIJSONLLine:
		var line harvest.TIJSONLLine
		if err := json.Unmarshal(env.Payload, &line); err != nil {
			return nil, err
		}
		pl := commit.TIJSONLRecordPayload{Line: line.Line}
		var lineEnv JSONLEnvelope
		if err := json.Unmarshal(pl.Line, &lineEnv); err != nil {
			return nil, fmt.Errorf("ti jsonl: %w", err)
		}
		return jsonlToIngest(lineEnv)

	case harvest.KindTIIoCRaw:
		var i tidomain.IOC
		if err := json.Unmarshal(env.Payload, &i); err != nil {
			return nil, err
		}
		ni, ok := tinormalize.NormalizeIOC(i)
		if !ok {
			return nil, nil
		}
		id := tinormalize.CanonicalID(ni)
		out, err := commit.NewEnvelope(commit.SourceTI, commit.KindTIIoC, commit.TIIoCIdempotencyKey(id), ni)
		return []*commit.Envelope{out}, err

	case harvest.KindTIReportRaw:
		var r tidomain.Report
		if err := json.Unmarshal(env.Payload, &r); err != nil {
			return nil, err
		}
		sid := tinormalize.ReportStableID(r.Link)
		out, err := commit.NewEnvelope(commit.SourceTI, commit.KindTIReport, commit.TIReportIdempotencyKey(sid), r)
		return []*commit.Envelope{out}, err

	case harvest.KindTICampaignRaw:
		var c tidomain.Campaign
		if err := json.Unmarshal(env.Payload, &c); err != nil {
			return nil, err
		}
		c = tinormalize.NormalizeCampaign(c)
		if c.ID == "" || c.Name == "" {
			return nil, fmt.Errorf("ti campaign requires id and name")
		}
		out, err := commit.NewEnvelope(commit.SourceTI, commit.KindTICampaign, commit.TICampaignIdempotencyKey(c.ID), c)
		return []*commit.Envelope{out}, err

	case harvest.KindTIClusterRaw:
		var cl tidomain.Cluster
		if err := json.Unmarshal(env.Payload, &cl); err != nil {
			return nil, err
		}
		cl = tinormalize.NormalizeCluster(cl)
		out, err := commit.NewEnvelope(commit.SourceTI, commit.KindTICluster, commit.TIClusterIdempotencyKey(cl.ID), cl)
		return []*commit.Envelope{out}, err

	case harvest.KindTIActorRaw:
		var a tidomain.Actor
		if err := json.Unmarshal(env.Payload, &a); err != nil {
			return nil, err
		}
		id := strings.TrimSpace(a.ID)
		if id == "" {
			id = tinormalize.ActorStableID(a.Name)
		}
		out, err := commit.NewEnvelope(commit.SourceTI, commit.KindTIActor, commit.TIActorIdempotencyKey(id), a)
		return []*commit.Envelope{out}, err

	default:
		return nil, fmt.Errorf("pipeline ti: unknown kind %q", env.Kind)
	}
}

func jsonlToIngest(lineEnv JSONLEnvelope) ([]*commit.Envelope, error) {
	var out []*commit.Envelope
	switch {
	case lineEnv.IOC != nil:
		ni, ok := tinormalize.NormalizeIOC(*lineEnv.IOC)
		if !ok {
			return nil, nil
		}
		id := tinormalize.CanonicalID(ni)
		e, err := commit.NewEnvelope(commit.SourceTI, commit.KindTIIoC, commit.TIIoCIdempotencyKey(id), ni)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	case lineEnv.Campaign != nil:
		c := tinormalize.NormalizeCampaign(*lineEnv.Campaign)
		e, err := commit.NewEnvelope(commit.SourceTI, commit.KindTICampaign, commit.TICampaignIdempotencyKey(c.ID), c)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	case lineEnv.Cluster != nil:
		cl := tinormalize.NormalizeCluster(*lineEnv.Cluster)
		e, err := commit.NewEnvelope(commit.SourceTI, commit.KindTICluster, commit.TIClusterIdempotencyKey(cl.ID), cl)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	case lineEnv.Actor != nil:
		a := *lineEnv.Actor
		id := strings.TrimSpace(a.ID)
		if id == "" {
			id = tinormalize.ActorStableID(a.Name)
		}
		e, err := commit.NewEnvelope(commit.SourceTI, commit.KindTIActor, commit.TIActorIdempotencyKey(id), a)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	case lineEnv.Report != nil:
		sid := tinormalize.ReportStableID(lineEnv.Report.Link)
		e, err := commit.NewEnvelope(commit.SourceTI, commit.KindTIReport, commit.TIReportIdempotencyKey(sid), *lineEnv.Report)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}
