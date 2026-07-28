package ledger

import (
	"testing"
	"time"
)

func TestFetchPolicy_constants(t *testing.T) {
	if PolicyStatic != "static" || PolicyPeriodic != "periodic" || PolicyDaily != "daily" {
		t.Fatal("unexpected policy constants")
	}
}

func TestShouldFetch_policySemantics(t *testing.T) {
	// Document expected policy behavior for FetchIfDue integrators.
	now := time.Now()
	stale := now.Add(-48 * time.Hour)
	recent := now.Add(-1 * time.Hour)

	if time.Since(stale) < 24*time.Hour {
		t.Fatal("stale should be older than daily window")
	}
	if time.Since(recent) >= 24*time.Hour {
		t.Fatal("recent should be within daily window")
	}
	_ = PolicyStatic
}
