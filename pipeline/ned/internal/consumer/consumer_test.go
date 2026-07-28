package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/config"
	"github.com/nats-io/nats.go"
)

type recordIngestPublisher struct {
	subject string
	envs    []*commit.Envelope
}

func (r *recordIngestPublisher) PublishJSON(_ context.Context, subject string, env *commit.Envelope) error {
	r.subject = subject
	r.envs = append(r.envs, env)
	return nil
}

func TestHandleScrapeMsg_invalidJSON(t *testing.T) {
	err := handleScrapeMsg(context.Background(), slog.Default(), &nats.Msg{Data: []byte("{")}, nil, config.Config{})
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestHandleScrapeMsg_invalidEnvelope(t *testing.T) {
	b, _ := json.Marshal(map[string]any{"source": "ti"})
	err := handleScrapeMsg(context.Background(), slog.Default(), &nats.Msg{Data: b}, nil, config.Config{})
	if err == nil {
		t.Fatal("expected validate error")
	}
}

func TestHandleScrapeMsg_tiIoCRaw_publishesIngest(t *testing.T) {
	raw := harvest.Envelope{
		SchemaVersion: harvest.CurrentSchemaVersion,
		Source:        harvest.SourceTI,
		Kind:          harvest.KindTIIoCRaw,
		ContentKey:    "ti:ioc:url:https://example.com",
		ScrapedAt:     "2026-05-15T12:00:00Z",
		Payload:       json.RawMessage(`{"type":"url","value":"https://example.com","source":"x"}`),
	}
	b, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	rec := &recordIngestPublisher{}
	if err := handleScrapeMsg(context.Background(), slog.Default(), &nats.Msg{Data: b}, rec, config.Config{}); err != nil {
		t.Fatal(err)
	}
	if len(rec.envs) == 0 {
		t.Fatal("expected published envelopes")
	}
	if rec.envs[0].Source != commit.SourceTI {
		t.Fatalf("source %s", rec.envs[0].Source)
	}
}
