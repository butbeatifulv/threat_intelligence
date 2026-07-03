package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/pkg/commit"
	engageevents "github.com/butbeautifulv/veil/pkg/engage/events"
	"github.com/butbeautifulv/veil/pkg/natsjet"
	"github.com/butbeautifulv/veil/pkg/observability"
	natsgo "github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"
)

// EnsureEngageEventsStream creates the ENGAGE_EVENTS stream for audit/finding subjects.
func EnsureEngageEventsStream(js natsgo.JetStreamContext) error {
	return natsjet.EnsureEngageEventsStream(js)
}

// RunEngageEventsConsumer pulls engage JetStream events and publishes ingest envelopes.
func RunEngageEventsConsumer(ctx context.Context, log *slog.Logger, natsURL, engageFilter, ingestToolRunSubject, ingestFindingSubject string) error {
	if log == nil {
		log = slog.Default()
	}
	pub, err := ConnectJetStreamAndStream(natsURL)
	if err != nil {
		return err
	}
	defer pub.Close()
	js := pub.conn.JS
	if err := EnsureEngageEventsStream(js); err != nil {
		return fmt.Errorf("engage events stream: %w", err)
	}
	if engageFilter == "" {
		engageFilter = "engage.events.>"
	}
	if ingestToolRunSubject == "" {
		ingestToolRunSubject = "ingest.engage.tool_run"
	}
	if ingestFindingSubject == "" {
		ingestFindingSubject = "ingest.engage.finding"
	}
	sub, err := js.PullSubscribe(engageFilter, "engage-events-bridge", natsgo.BindStream(natsjet.StreamEngageEvents))
	if err != nil {
		return fmt.Errorf("pull subscribe: %w", err)
	}
	defer sub.Unsubscribe()
	return natsjet.RunPullLoop(ctx, log, sub, natsjet.PullLoopOpts{
		Batch:              10,
		MaxWait:            2 * time.Second,
		ErrOnFetch:         true,
		ReturnContextError: true,
	}, func(ctx context.Context, m *natsgo.Msg) error {
		if err := handleEngageMsg(ctx, pub, ingestToolRunSubject, ingestFindingSubject, m); err != nil {
			log.Warn("engage event", slog.Any("err", err))
			return err
		}
		return nil
	})
}

func handleEngageMsg(ctx context.Context, pub *JetStreamPublisher, ingestToolRunSubject, ingestFindingSubject string, m *natsgo.Msg) error {
	start := time.Now()
	status := "ok"
	subject := m.Subject
	ctx, span := observability.StartWorkerSpan(ctx, "engage-events", "engage-events.process",
		attribute.String("nats.subject", subject),
	)
	defer func() {
		observability.RecordNatsMessage("engage-events", subject, status, time.Since(start).Seconds())
		span.End()
	}()

	filterCtx, filterSpan := observability.StartSpan(ctx, "engage-events.filter")
	var err error
	if strings.Contains(m.Subject, ".finding") {
		err = handleEngageFinding(filterCtx, pub, ingestFindingSubject, m)
	} else {
		err = handleEngageToolRun(filterCtx, pub, ingestToolRunSubject, m)
	}
	filterSpan.End()
	if err != nil {
		status = "error"
		observability.RecordError(span, err)
		return err
	}
	return nil
}

func handleEngageToolRun(ctx context.Context, pub *JetStreamPublisher, ingestSubject string, m *natsgo.Msg) error {
	var wire engageevents.AuditEvent
	if err := json.Unmarshal(m.Data, &wire); err != nil {
		return err
	}
	if strings.TrimSpace(wire.Tool) == "" {
		return fmt.Errorf("empty tool")
	}
	at := wire.At.UTC()
	if at.IsZero() {
		at = time.Now().UTC()
	}
	atStr := at.Format(time.RFC3339)
	payload, err := json.Marshal(commit.EngageToolRunPayload{
		Tool:    wire.Tool,
		Target:  wire.Target,
		Subject: wire.Subject,
		Success: wire.Success,
		At:      atStr,
	})
	if err != nil {
		return err
	}
	env := &commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceEngage,
		Kind:           commit.KindEngageToolRun,
		IdempotencyKey: commit.EngageToolRunIdempotencyKey(wire.Tool, wire.Target, atStr),
		Payload:        payload,
	}
	return pub.PublishJSON(ctx, ingestSubject, env)
}

func handleEngageFinding(ctx context.Context, pub *JetStreamPublisher, ingestSubject string, m *natsgo.Msg) error {
	var wire engageevents.FindingEvent
	if err := json.Unmarshal(m.Data, &wire); err != nil {
		return err
	}
	if strings.TrimSpace(wire.Tool) == "" || strings.TrimSpace(wire.Title) == "" {
		return fmt.Errorf("empty tool or title")
	}
	payload, err := json.Marshal(commit.EngageFindingPayload{
		Tool:        wire.Tool,
		Target:      wire.Target,
		Title:       wire.Title,
		Severity:    wire.Severity,
		Description: wire.Description,
	})
	if err != nil {
		return err
	}
	env := &commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceEngage,
		Kind:           commit.KindEngageFinding,
		IdempotencyKey: commit.EngageFindingIdempotencyKey(wire.Tool, wire.Target, wire.Title),
		Payload:        payload,
	}
	return pub.PublishJSON(ctx, ingestSubject, env)
}
