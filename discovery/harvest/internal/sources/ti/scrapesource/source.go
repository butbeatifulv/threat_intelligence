// Package scrapesource registers the ti scrape source with harvest/internal/factory.
package scrapesource

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/factory"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/ti/internal/scrapepub"
	"github.com/butbeautifulv/veil/discovery/harvest/internal/sources/ti/internal/usecase"
)

func init() {
	factory.Register("ti", func() factory.Source { return &Source{} })
}

// Source scrapes public TI feeds and optional local JSONL.
type Source struct{}

func (s *Source) Name() string { return "ti" }

func (s *Source) Policy() factory.FetchPolicy { return factory.PolicyDaily }

func (s *Source) Run(ctx context.Context, deps *factory.ScrapeDeps) error {
	pub, err := deps.Publisher("ti")
	if err != nil {
		return err
	}
	repo := scrapepub.NewFromRaw(pub)
	runner := usecase.NewRunner(repo, deps.Log, deps.Feeds, deps.Ledger)

	if env := strings.TrimSpace(os.Getenv("TI_FEEDS")); env != "" {
		kinds := strings.Split(env, ",")
		if err := runner.Run(ctx, kinds); err != nil {
			return err
		}
	}

	if path := strings.TrimSpace(os.Getenv("TI_JSONL_FILE")); path != "" {
		if err := ingestJSONLFile(ctx, repo, path); err != nil {
			return err
		}
	}

	if strings.TrimSpace(os.Getenv("TI_FEEDS")) == "" && strings.TrimSpace(os.Getenv("TI_JSONL_FILE")) == "" {
		deps.Log.Info("ti: nothing to do; set TI_FEEDS and/or TI_JSONL_FILE")
	}
	return nil
}

func ingestJSONLFile(ctx context.Context, repo *scrapepub.Publisher, path string) error {
	f, err := os.Open(path) // #nosec G304 -- path is the TI_JSONL_FILE deploy-time env var, operator-controlled, not request/network input
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for sc.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		if err := repo.PublishJSONLLine(ctx, line); err != nil {
			return err
		}
	}
	return sc.Err()
}
