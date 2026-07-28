package domain

import (
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
)

func TestSourceValid(t *testing.T) {
	for _, s := range AllSources() {
		if !s.Valid() {
			t.Fatalf("%q should be valid", s)
		}
	}
	if Source("unknown").Valid() {
		t.Fatal("unknown source should be invalid")
	}
}

func TestAllSourcesMatchesHarvestCommit(t *testing.T) {
	harvestWant := []string{
		harvest.SourceSBOM,
		harvest.SourceCoderules,
		harvest.SourceNuclei,
		harvest.SourceTI,
		harvest.SourceVuln,
		harvest.SourceLola,
		harvest.SourceDS,
		harvest.SourceBrowser,
	}
	commitWant := []string{
		commit.SourceSBOM,
		commit.SourceCoderules,
		commit.SourceNuclei,
		commit.SourceTI,
		commit.SourceVuln,
		commit.SourceLola,
		commit.SourceDS,
		commit.SourceEngage,
	}
	registry := make(map[Source]struct{})
	for _, s := range AllSources() {
		registry[s] = struct{}{}
	}
	for _, wire := range harvestWant {
		if _, ok := registry[Source(wire)]; !ok {
			t.Fatalf("registry missing harvest source %q", wire)
		}
		if Source(wire).String() != wire {
			t.Fatalf("Source(%q).String() mismatch", wire)
		}
	}
	for _, wire := range commitWant {
		if _, ok := registry[Source(wire)]; !ok {
			t.Fatalf("registry missing commit source %q", wire)
		}
	}
	// harvest-only
	if Source(harvest.SourceBrowser).String() != harvest.SourceBrowser {
		t.Fatal("browser wire mismatch")
	}
	// commit-only
	if Source(commit.SourceEngage).String() != commit.SourceEngage {
		t.Fatal("engage wire mismatch")
	}
}
