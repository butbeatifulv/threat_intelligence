package natsjet

import "github.com/nats-io/nats.go"

const (
	StreamScrape       = "SCRAPE"
	StreamIngest       = "INGEST"
	StreamEngageEvents = "ENGAGE_EVENTS"
)

// EnsureIngestStream creates or widens INGEST (ingest.>).
func EnsureIngestStream(js nats.JetStreamContext) error {
	return EnsureStream(js, StreamIngest, []string{"ingest.>"})
}

// EnsureScrapeStream creates or widens SCRAPE (scrape.>).
func EnsureScrapeStream(js nats.JetStreamContext) error {
	return EnsureStream(js, StreamScrape, []string{"scrape.>"})
}

// EnsureEngageEventsStream creates or widens ENGAGE_EVENTS (engage.events.>).
func EnsureEngageEventsStream(js nats.JetStreamContext) error {
	return EnsureStream(js, StreamEngageEvents, []string{"engage.events.>"})
}

var ensureScrapeStreamFn = EnsureScrapeStream

// EnsureScrapeAndIngest ensures SCRAPE and INGEST (pipeline worker).
func EnsureScrapeAndIngest(js nats.JetStreamContext) error {
	if err := ensureScrapeStreamFn(js); err != nil {
		return err
	}
	return EnsureIngestStream(js)
}
