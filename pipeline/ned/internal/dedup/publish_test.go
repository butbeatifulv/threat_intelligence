package dedup

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
	tidomain "github.com/butbeautifulv/veil/pkg/ti/domain"
)

type recordPublisher struct {
	subject string
	envs    []*commit.Envelope
}

func (r *recordPublisher) PublishJSON(_ context.Context, subject string, env *commit.Envelope) error {
	r.subject = subject
	r.envs = append(r.envs, env)
	return nil
}

func TestProcessScrapeMessage_tiIoCRaw_publishesIngest(t *testing.T) {
	raw := tidomain.IOC{Type: tidomain.IOCURL, Value: "https://evil.com/x", Source: "test"}
	payload, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	env := &harvest.Envelope{
		SchemaVersion: harvest.CurrentSchemaVersion,
		Source:        harvest.SourceTI,
		Kind:          harvest.KindTIIoCRaw,
		ContentKey:    "ti:ioc:url:https://evil.com/x",
		ScrapedAt:     "2026-05-15T12:00:00Z",
		Payload:       payload,
	}
	rec := &recordPublisher{}
	if err := ProcessScrapeMessage(context.Background(), rec, "ingest.ti.events", env); err != nil {
		t.Fatal(err)
	}
	if rec.subject != "ingest.ti.events" {
		t.Fatalf("subject %q", rec.subject)
	}
	if len(rec.envs) != 1 {
		t.Fatalf("published %d envelopes", len(rec.envs))
	}
	if rec.envs[0].Kind != commit.KindTIIoC {
		t.Fatalf("kind %s", rec.envs[0].Kind)
	}
	if err := rec.envs[0].Validate(); err != nil {
		t.Fatal(err)
	}
}
