package query

import (
	"strings"
	"testing"
)

func TestNormalizeEngageHost(t *testing.T) {
	got := NormalizeEngageHost("https://Example.COM/path")
	if got != "example.com" {
		t.Fatalf("got %q", got)
	}
}

func TestSeedMatchByID_includesEngageTarget(t *testing.T) {
	if !strings.Contains(seedMatchByID, "EngageTarget") {
		t.Fatal("seedMatchByID missing EngageTarget")
	}
}
