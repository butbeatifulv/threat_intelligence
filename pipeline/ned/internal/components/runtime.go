package components

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/butbeautifulv/veil/pipeline/connector/nats"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/config"
	"github.com/butbeautifulv/veil/pipeline/ned/internal/consumer"
	natsgo "github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
)

// Runtime holds NATS scrape subscription and ingest publisher for the NED worker.
type Runtime struct {
	cfg    config.Config
	log    *slog.Logger
	nc     *natsgo.Conn
	sub    *natsgo.Subscription
	pub    *nats.JetStreamPublisher
}

// Init wires NATS scrape consumer and ingest publisher.
func Init(ctx context.Context, cfg config.Config, log *slog.Logger) (*Runtime, error) {
	_ = ctx
	nc, err := natsgo.Connect(cfg.NATSURL)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		_ = nc.Drain()
		return nil, fmt.Errorf("jetstream: %w", err)
	}
	if err := nats.EnsureBothStreams(js); err != nil {
		_ = nc.Drain()
		return nil, fmt.Errorf("streams: %w", err)
	}
	pub, err := nats.ConnectJetStream(cfg.NATSURL)
	if err != nil {
		_ = nc.Drain()
		return nil, err
	}
	sub, err := js.PullSubscribe(cfg.ScrapeSubject, cfg.ScrapeDurable, natsgo.BindStream(cfg.ScrapeStream))
	if err != nil {
		pub.Close()
		_ = nc.Drain()
		return nil, fmt.Errorf("pull subscribe scrape stream=%s: %w", cfg.ScrapeStream, err)
	}
	return &Runtime{cfg: cfg, log: log, nc: nc, sub: sub, pub: pub}, nil
}

// Run starts the pull consumer until ctx is canceled.
func (r *Runtime) Run(ctx context.Context) error {
	r.log.Info("pipeline-worker started",
		slog.String("nats", r.cfg.NATSURL),
		slog.String("scrape_stream", r.cfg.ScrapeStream),
		slog.String("scrape_subject", r.cfg.ScrapeSubject),
		slog.String("ingest_publish", r.cfg.IngestPublish),
	)
	return consumer.RunPullLoop(ctx, r.log, r.sub, r.cfg.Batch, r.cfg.MaxWait, r.pub, r.cfg)
}

// Shutdown drains NATS resources.
func (r *Runtime) Shutdown() {
	if r.sub != nil {
		_ = r.sub.Unsubscribe()
	}
	if r.pub != nil {
		r.pub.Close()
	}
	if r.nc != nil {
		_ = r.nc.Drain()
	}
}

// Run is the top-level worker lifecycle.
func Run(rootCtx context.Context, log *slog.Logger) error {
	cfg := config.Load()
	rt, err := Init(rootCtx, cfg, log)
	if err != nil {
		return err
	}
	defer rt.Shutdown()

	eg, ctx := errgroup.WithContext(rootCtx)
	eg.Go(func() error {
		return rt.Run(ctx)
	})
	return eg.Wait()
}
