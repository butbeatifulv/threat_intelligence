package factory

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	connats "github.com/butbeautifulv/veil/discovery/connector/nats"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/feeds"
)

// RunOptions configures a scrape-worker run.
type RunOptions struct {
	SourceNames []string
	NATSURL     string
	CacheDir    string
	Log         *slog.Logger
}

// Run connects shared deps and runs selected sources via the registry.
func Run(ctx context.Context, opts RunOptions) error {
	if opts.Log == nil {
		opts.Log = slog.Default()
	}
	natsURL := strings.TrimSpace(opts.NATSURL)
	if natsURL == "" {
		natsURL = envOr("NATS_URL", "nats://localhost:4222")
	}

	sources, err := SourcesFor(opts.SourceNames)
	if err != nil {
		return err
	}

	jsPub, err := connats.ConnectJetStreamAndStream(natsURL)
	if err != nil {
		return fmt.Errorf("nats: %w", err)
	}
	defer jsPub.Close()

	led, err := feeds.OpenLedgerFromEnv(ctx)
	if err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	if led != nil {
		defer func() { _ = led.Close() }()
	}

	cacheDir := strings.TrimSpace(opts.CacheDir)
	if cacheDir == "" {
		cacheDir = strings.TrimSpace(os.Getenv("SCRAPE_CACHE_DIR"))
	}
	feedsClient := feeds.NewClient(cacheDir, opts.Log)

	publishers := make(map[string]RawPublisher, len(sources))
	for _, src := range sources {
		name := src.Name()
		srcConst, ok := scrapeSourceConstant(name)
		if !ok {
			return fmt.Errorf("scrape: no harvest source constant for %q", name)
		}
		subj := subjectForSource(name)
		publishers[name] = connats.NewDomainPublisher(jsPub, srcConst, subj)
	}

	deps := &ScrapeDeps{
		Ledger:     led,
		Feeds:      feedsClient,
		Log:        opts.Log,
		Publishers: publishers,
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err = NewRegistry(sources...).RunAll(ctx, deps)
	if err != nil {
		opts.Log.Info("scrape-worker finished", slog.String("error", err.Error()))
		return err
	}
	opts.Log.Info("scrape-worker finished")
	return nil
}
