package ids

import (
	"testing"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
)

func TestActorStableID_stable(t *testing.T) {
	a := ActorStableID("APT29")
	b := ActorStableID(" apt29 ")
	if a == "" || a != b {
		t.Fatalf("a=%q b=%q", a, b)
	}
}

func TestReportStableID_stable(t *testing.T) {
	link := "https://example.com/report/1"
	a := ReportStableID(link)
	b := ReportStableID(" " + link + " ")
	if a == "" || a != b {
		t.Fatalf("a=%q b=%q", a, b)
	}
}

func TestCanonicalIOCID_matchesNodeID(t *testing.T) {
	ioc := domain.IOC{Type: domain.IOCIP, Value: "203.0.113.1"}
	if CanonicalIOCID(ioc) != ioc.NodeID() {
		t.Fatal("mismatch")
	}
}
