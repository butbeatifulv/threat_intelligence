package natsjet

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func TestEnsureScrapeAndIngest_scrapeError(t *testing.T) {
	old := ensureScrapeStreamFn
	defer func() { ensureScrapeStreamFn = old }()
	want := errors.New("scrape fail")
	ensureScrapeStreamFn = func(nats.JetStreamContext) error { return want }
	err := EnsureScrapeAndIngest(jetStreamForTest(t))
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestEnsureScrapeAndIngest_bothStreams(t *testing.T) {
	js := jetStreamForTest(t)
	if err := EnsureScrapeAndIngest(js); err != nil {
		t.Fatal(err)
	}
	if _, err := js.StreamInfo(StreamScrape); err != nil {
		t.Fatal(err)
	}
	if _, err := js.StreamInfo(StreamIngest); err != nil {
		t.Fatal(err)
	}
}

func TestRunPullLoop_contextError(t *testing.T) {
	_, url := startTestNATSPull(t)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()
	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureStream(js, "PULL_CTX", []string{"pull.ctx.>"}); err != nil {
		t.Fatal(err)
	}
	_, err = js.AddConsumer("PULL_CTX", &nats.ConsumerConfig{
		Durable: "d", AckPolicy: nats.AckExplicitPolicy,
	})
	if err != nil {
		t.Fatal(err)
	}
	sub, err := js.PullSubscribe("pull.ctx.>", "d", nats.Bind("PULL_CTX", "d"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = RunPullLoop(ctx, slog.Default(), sub, PullLoopOpts{
		Batch: 1, MaxWait: 50 * time.Millisecond, ReturnContextError: true, StopLog: "stop",
	}, func(context.Context, *nats.Msg) error { return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestRunPullLoop_handlerNakNoDelay(t *testing.T) {
	_, url := startTestNATSPull(t)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()
	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureStream(js, "PULL_NAK0", []string{"pull.nak0.>"}); err != nil {
		t.Fatal(err)
	}
	if _, err := js.Publish("pull.nak0.>", []byte("x")); err != nil {
		t.Fatal(err)
	}
	sub, err := js.PullSubscribe("pull.nak0.>", "pull-nak0", nats.BindStream("PULL_NAK0"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	_ = RunPullLoop(ctx, slog.Default(), sub, PullLoopOpts{
		Batch: 1, MaxWait: 40 * time.Millisecond,
	}, func(context.Context, *nats.Msg) error { return errors.New("fail") })
}

func TestRunPullLoop_handlerErrorNak(t *testing.T) {
	_, url := startTestNATSPull(t)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()
	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureStream(js, "PULL_NAK", []string{"pull.nak.>"}); err != nil {
		t.Fatal(err)
	}
	_, err = js.AddConsumer("PULL_NAK", &nats.ConsumerConfig{
		Durable: "dn", AckPolicy: nats.AckExplicitPolicy,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := js.Publish("pull.nak.>", []byte("x")); err != nil {
		t.Fatal(err)
	}
	sub, err := js.PullSubscribe("pull.nak.>", "dn", nats.Bind("PULL_NAK", "dn"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
	defer cancel()
	_ = RunPullLoop(ctx, slog.Default(), sub, PullLoopOpts{
		Batch: 1, MaxWait: 50 * time.Millisecond, NakDelay: 10 * time.Millisecond,
	}, func(context.Context, *nats.Msg) error {
		return errors.New("handler fail")
	})
}

func TestRunPullLoop_fetchErrorBackoff(t *testing.T) {
	old := pullFetchBackoff
	pullFetchBackoff = 5 * time.Millisecond
	defer func() { pullFetchBackoff = old }()

	_, url := startTestNATSPull(t)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()
	sub, err := nc.SubscribeSync("pull.backoff.>")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	if err := RunPullLoop(ctx, slog.Default(), sub, PullLoopOpts{
		Batch: 1, MaxWait: 10 * time.Millisecond,
	}, func(context.Context, *nats.Msg) error { return nil }); err != nil {
		t.Fatal(err)
	}
}

func TestRunPullLoop_fetchErrorCancelNoReturn(t *testing.T) {
	_, url := startTestNATSPull(t)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()
	sub, err := nc.SubscribeSync("pull.fetch.nil.>")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(25 * time.Millisecond)
		cancel()
	}()
	if err := RunPullLoop(ctx, slog.Default(), sub, PullLoopOpts{
		Batch: 1, MaxWait: 15 * time.Millisecond,
	}, func(context.Context, *nats.Msg) error { return nil }); err != nil {
		t.Fatalf("got %v", err)
	}
}

func TestRunPullLoop_fetchErrorReturnContext(t *testing.T) {
	_, url := startTestNATSPull(t)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()
	sub, err := nc.SubscribeSync("pull.fetch.ctx.>")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()
	err = RunPullLoop(ctx, slog.Default(), sub, PullLoopOpts{
		Batch: 1, MaxWait: 20 * time.Millisecond, ReturnContextError: true,
	}, func(context.Context, *nats.Msg) error { return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestRunPullLoop_ackWarn(t *testing.T) {
	_, url := startTestNATSPull(t)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()
	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureStream(js, "PULL_ACK", []string{"pull.ack.>"}); err != nil {
		t.Fatal(err)
	}
	if _, err := js.Publish("pull.ack.>", []byte("x")); err != nil {
		t.Fatal(err)
	}
	sub, err := js.PullSubscribe("pull.ack.>", "pull-ack", nats.BindStream("PULL_ACK"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	_ = RunPullLoop(ctx, slog.Default(), sub, PullLoopOpts{Batch: 1, MaxWait: 50 * time.Millisecond},
		func(_ context.Context, m *nats.Msg) error {
			_ = m.Ack()
			return nil
		})
}

func TestRunPullLoop_stopLogAndFetchWarn(t *testing.T) {
	_, url := startTestNATSPull(t)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()
	sub, err := nc.SubscribeSync("pull.warn.>")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = RunPullLoop(ctx, slog.Default(), sub, PullLoopOpts{
		Batch: 1, MaxWait: 10 * time.Millisecond, StopLog: "stopped",
	}, func(context.Context, *nats.Msg) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunPullLoop_fetchErrorErrOnFetch(t *testing.T) {
	_, url := startTestNATSPull(t)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatal(err)
	}
	defer nc.Drain()
	sub, err := nc.SubscribeSync("not.jetstream")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = RunPullLoop(ctx, slog.Default(), sub, PullLoopOpts{
		Batch: 1, MaxWait: 50 * time.Millisecond, ErrOnFetch: true,
	}, func(context.Context, *nats.Msg) error { return nil })
	if err == nil {
		t.Fatal("expected fetch error")
	}
}
