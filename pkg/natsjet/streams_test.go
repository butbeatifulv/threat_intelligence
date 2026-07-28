package natsjet

import (
	"testing"

	"github.com/nats-io/nats.go"
)

func TestEnsurePlatformStreams(t *testing.T) {
	js := jetStreamForTest(t)
	for _, fn := range []struct {
		name string
		run  func(nats.JetStreamContext) error
	}{
		{"scrape", EnsureScrapeStream},
		{"ingest", EnsureIngestStream},
		{"engage_events", EnsureEngageEventsStream},
	} {
		t.Run(fn.name, func(t *testing.T) {
			if err := fn.run(js); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEnsureScrapeAndIngest(t *testing.T) {
	js := jetStreamForTest(t)
	if err := EnsureScrapeAndIngest(js); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{StreamScrape, StreamIngest} {
		if _, err := js.StreamInfo(name); err != nil {
			t.Fatalf("stream %s: %v", name, err)
		}
	}
}

func TestStreamConstants(t *testing.T) {
	if StreamScrape == "" || StreamIngest == "" || StreamEngageEvents == "" {
		t.Fatal("stream name constants must be non-empty")
	}
}
