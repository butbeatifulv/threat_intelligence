package natsjet

import (
	"context"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
	natsgo "github.com/nats-io/nats.go"
)

func TestPublishHarvest_errorsWithoutNATS(t *testing.T) {
	t.Parallel()
	c := &Conn{}
	ctx := context.Background()

	tests := []struct {
		name    string
		source  string
		kind    string
		key     string
		payload any
		wantSub string
	}{
		{
			name:    "marshal error",
			source:  harvest.SourceTI,
			kind:    harvest.KindTIIoCRaw,
			key:     "ti:ioc:1",
			payload: make(chan int),
			wantSub: "unsupported type",
		},
		{
			name:    "empty source",
			source:  "",
			kind:    harvest.KindTIIoCRaw,
			key:     "ti:ioc:1",
			payload: map[string]string{"id": "1"},
			wantSub: "empty source",
		},
		{
			name:    "empty content key",
			source:  harvest.SourceTI,
			kind:    harvest.KindTIIoCRaw,
			key:     "",
			payload: map[string]string{"id": "1"},
			wantSub: "content_key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PublishHarvest(ctx, c, "scrape.ti.events", tt.source, tt.kind, tt.key, tt.payload)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Fatalf("err = %v, want substring %q", err, tt.wantSub)
			}
		})
	}
}

func TestPublishHarvestEnvelope_validateBeforePublish(t *testing.T) {
	t.Parallel()
	env := &harvest.Envelope{
		SchemaVersion: harvest.CurrentSchemaVersion,
		Source:        harvest.SourceTI,
		Kind:          harvest.KindTIIoCRaw,
		ContentKey:    "ti:ioc:1",
		ScrapedAt:     "2020-01-01T00:00:00Z",
		Payload:       nil,
	}
	err := PublishHarvestEnvelope(context.Background(), &Conn{}, "scrape.ti.events", env)
	if err == nil || !strings.Contains(err.Error(), "payload") {
		t.Fatalf("err = %v", err)
	}
}

func TestPublishCommit_errorsWithoutNATS(t *testing.T) {
	t.Parallel()
	c := &Conn{}
	ctx := context.Background()

	tests := []struct {
		name    string
		source  string
		kind    string
		key     string
		payload any
		wantSub string
	}{
		{
			name:    "marshal error",
			source:  commit.SourceEngage,
			kind:    commit.KindEngageToolRun,
			key:     "engage:run:1",
			payload: make(chan int),
			wantSub: "unsupported type",
		},
		{
			name:    "empty idempotency key",
			source:  commit.SourceEngage,
			kind:    commit.KindEngageToolRun,
			key:     "",
			payload: map[string]any{"tool": "nmap"},
			wantSub: "idempotency_key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PublishCommit(ctx, c, "ingest.engage.events", tt.source, tt.kind, tt.key, tt.payload)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Fatalf("err = %v, want substring %q", err, tt.wantSub)
			}
		})
	}
}

func connForTest(t *testing.T) *Conn {
	t.Helper()
	opts := &server.Options{JetStream: true, StoreDir: t.TempDir(), Port: -1}
	srv := test.RunServer(opts)
	t.Cleanup(srv.Shutdown)
	nc, err := natsgo.Connect(srv.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = nc.Drain() })
	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	return &Conn{NC: nc, JS: js}
}

func TestPublishHarvest_contentKeyDedup(t *testing.T) {
	c := connForTest(t)
	if err := EnsureScrapeStream(c.JS); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	subject := "scrape.ti.dedup"
	key := "ti:ioc:dedup:1.2.3.4"
	pl := map[string]string{"value": "1.2.3.4"}
	if err := PublishHarvest(ctx, c, subject, harvest.SourceTI, harvest.KindTIIoCRaw, key, pl); err != nil {
		t.Fatal(err)
	}
	if err := PublishHarvest(ctx, c, subject, harvest.SourceTI, harvest.KindTIIoCRaw, key, pl); err != nil {
		t.Fatal(err)
	}
	info, err := c.JS.StreamInfo(StreamScrape)
	if err != nil {
		t.Fatal(err)
	}
	if info.State.Msgs != 1 {
		t.Fatalf("stream msgs = %d, want 1 (dedup by content_key msg id)", info.State.Msgs)
	}
}

func TestPublishCommit_idempotencyKeyDedup(t *testing.T) {
	c := connForTest(t)
	if err := EnsureIngestStream(c.JS); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	subject := "ingest.engage.dedup"
	key := "engage:run:nmap:host:2020-01-01T00:00:00Z"
	pl := map[string]any{"tool": "nmap", "target": "host"}
	if err := PublishCommit(ctx, c, subject, commit.SourceEngage, commit.KindEngageToolRun, key, pl); err != nil {
		t.Fatal(err)
	}
	if err := PublishCommit(ctx, c, subject, commit.SourceEngage, commit.KindEngageToolRun, key, pl); err != nil {
		t.Fatal(err)
	}
	info, err := c.JS.StreamInfo(StreamIngest)
	if err != nil {
		t.Fatal(err)
	}
	if info.State.Msgs != 1 {
		t.Fatalf("stream msgs = %d, want 1 (dedup by idempotency_key msg id)", info.State.Msgs)
	}
}
