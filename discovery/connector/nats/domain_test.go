package nats

import (
	"testing"

	"github.com/butbeautifulv/veil/pkg/harvest"
)

func TestNewDomainPublisher_trimsSubject(t *testing.T) {
	dp := NewDomainPublisher(nil, harvest.SourceDS, "  scrape.ds.events  ")
	if dp.Subject != "scrape.ds.events" {
		t.Fatalf("subject = %q", dp.Subject)
	}
	if dp.Source != harvest.SourceDS {
		t.Fatalf("source = %q", dp.Source)
	}
}
