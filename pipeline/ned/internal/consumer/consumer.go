package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/butbeautifulv/veil/pkg/harvest"
	"github.com/butbeautifulv/veil/pkg/natsjet"
	"github.com/butbeautifulv/veil/pkg/observability"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/config"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/dedup"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"
)

// RunPullLoop consumes scrape JetStream messages until ctx is canceled.
func RunPullLoop(ctx context.Context, log *slog.Logger, sub *nats.Subscription, batch int, maxWait time.Duration, pub dedup.IngestPublisher, cfg config.Config) error {
	return natsjet.RunPullLoop(ctx, log, sub, natsjet.PullLoopOpts{
		Batch:    batch,
		MaxWait:  maxWait,
		NakDelay: 2 * time.Second,
		StopLog:  "pipeline consumer stopped",
	}, func(ctx context.Context, m *nats.Msg) error {
		return handleScrapeMsg(ctx, log, m, pub, cfg)
	})
}

func handleScrapeMsg(ctx context.Context, log *slog.Logger, m *nats.Msg, pub dedup.IngestPublisher, cfg config.Config) error {
	start := time.Now()
	status := "ok"
	subject := m.Subject
	ctx, span := observability.StartWorkerSpan(ctx, "pipeline", "pipeline.process_message",
		attribute.String("nats.subject", subject),
	)
	defer func() {
		observability.RecordNatsMessage("pipeline", subject, status, time.Since(start).Seconds())
		span.End()
	}()

	var env harvest.Envelope
	_, normSpan := observability.StartSpan(ctx, "pipeline.normalize")
	err := json.Unmarshal(m.Data, &env)
	if err != nil {
		normSpan.End()
		status = "error"
		observability.RecordError(span, err)
		return fmt.Errorf("decode harvest: %w", err)
	}
	if err := env.Validate(); err != nil {
		normSpan.End()
		status = "error"
		observability.RecordError(span, err)
		return err
	}
	normSpan.End()

	subj := cfg.IngestSubjectFor(env.Source)
	pubCtx, pubSpan := observability.StartSpan(ctx, "pipeline.publish.ingest",
		attribute.String("envelope.source", env.Source),
		attribute.String("envelope.kind", env.Kind),
		attribute.String("ingest.subject", subj),
	)
	err = dedup.ProcessScrapeMessage(pubCtx, pub, subj, &env)
	pubSpan.End()
	if err != nil {
		status = "error"
		observability.RecordError(span, err)
		return err
	}
	log.Debug("pipeline processed", slog.String("source", env.Source), slog.String("kind", env.Kind))
	return nil
}
