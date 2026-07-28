package natsjet

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/butbeautifulv/veil/pkg/observability"
	"github.com/nats-io/nats.go"
)

var pullFetchBackoff = time.Second

// PullLoopOpts configures a JetStream pull consumer loop.
type PullLoopOpts struct {
	Batch    int
	MaxWait  time.Duration
	NakDelay time.Duration // >0: NakWithDelay; 0: Nak on handler error
	// ErrOnFetch aborts on non-timeout fetch errors (engage-events bridge).
	ErrOnFetch bool
	// ReturnContextError makes ctx cancellation return ctx.Err() instead of nil.
	ReturnContextError bool
	StopLog            string // logged when ctx is canceled; empty skips
}

// RunPullLoop fetches batches until ctx is canceled.
func RunPullLoop(ctx context.Context, log *slog.Logger, sub *nats.Subscription, opts PullLoopOpts, handle func(context.Context, *nats.Msg) error) error {
	if log == nil {
		log = slog.Default()
	}
	for {
		select {
		case <-ctx.Done():
			if opts.StopLog != "" {
				log.Info(opts.StopLog)
			}
			if opts.ReturnContextError {
				return ctx.Err()
			}
			return nil
		default:
		}
		msgs, err := sub.Fetch(opts.Batch, nats.MaxWait(opts.MaxWait))
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				continue
			}
			if opts.ErrOnFetch {
				return err
			}
			log.Warn("fetch", slog.String("err", err.Error()))
			select {
			case <-ctx.Done():
				if opts.ReturnContextError {
					return ctx.Err()
				}
				return nil
			case <-time.After(pullFetchBackoff):
			}
			continue
		}
		for _, m := range msgs {
			msgCtx := observability.ExtractTraceContext(ctx, m)
			handled := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Warn("message handler panicked", slog.Any("panic", r))
						if opts.NakDelay > 0 {
							_ = m.NakWithDelay(opts.NakDelay)
						} else {
							_ = m.Nak()
						}
						handled = true
					}
				}()
				if err := handle(msgCtx, m); err != nil {
					log.Warn("message", slog.String("err", err.Error()))
					if opts.NakDelay > 0 {
						_ = m.NakWithDelay(opts.NakDelay)
					} else {
						_ = m.Nak()
					}
					handled = true
					return
				}
			}()
			if handled {
				continue
			}
			if err := m.Ack(); err != nil {
				log.Warn("ack", slog.String("err", err.Error()))
			}
		}
	}
}
