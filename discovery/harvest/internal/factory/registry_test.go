package factory

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/butbeautifulv/veil/discovery/harvest/internal/ledger"
)

type mockSource struct {
	name    string
	policy  FetchPolicy
	runErr  error
	runCall int
}

func (m *mockSource) Name() string               { return m.name }
func (m *mockSource) Policy() FetchPolicy        { return m.policy }
func (m *mockSource) Run(ctx context.Context, deps *ScrapeDeps) error {
	m.runCall++
	return m.runErr
}

func TestRegistry_RunAll_orderAndError(t *testing.T) {
	t.Setenv("SCRAPE_FAIL_FAST", "1")
	log := slog.Default()
	a := &mockSource{name: "a", policy: PolicyPeriodic}
	b := &mockSource{name: "b", policy: PolicyDaily, runErr: errors.New("boom")}
	deps := &ScrapeDeps{Log: log, Publishers: map[string]RawPublisher{}}

	reg := NewRegistry(a, b)
	err := reg.RunAll(context.Background(), deps)
	if err == nil || err.Error() != "b: boom" {
		t.Fatalf("err = %v", err)
	}
	if a.runCall != 1 || b.runCall != 1 {
		t.Fatalf("run calls: a=%d b=%d", a.runCall, b.runCall)
	}
}

func TestSourcesFor_registered(t *testing.T) {
	Register("ds", func() Source {
		return &mockSource{name: "ds", policy: PolicyPeriodic}
	})
	srcs, err := SourcesFor([]string{"ds"})
	if err != nil {
		t.Fatal(err)
	}
	if len(srcs) != 1 || srcs[0].Name() != "ds" {
		t.Fatalf("unexpected: %+v", srcs[0])
	}
}

func TestSourcesFor_unimplemented(t *testing.T) {
	_, err := SourcesFor([]string{"not-a-source"})
	if err == nil {
		t.Fatal("expected error for unknown source")
	}
}

func TestSourcesFor_sbom(t *testing.T) {
	Register("sbom", func() Source { return &mockSource{name: "sbom", policy: PolicyPeriodic} })
	srcs, err := SourcesFor([]string{"sbom"})
	if err != nil {
		t.Fatal(err)
	}
	if len(srcs) != 1 || srcs[0].Name() != "sbom" {
		t.Fatalf("unexpected: %+v", srcs)
	}
}

func TestSourcesFor_vulnLola(t *testing.T) {
	Register("vuln", func() Source { return &mockSource{name: "vuln", policy: PolicyPeriodic} })
	Register("lola", func() Source { return &mockSource{name: "lola", policy: PolicyPeriodic} })
	srcs, err := SourcesFor([]string{"vuln", "lola"})
	if err != nil {
		t.Fatal(err)
	}
	if len(srcs) != 2 || srcs[0].Name() != "vuln" || srcs[1].Name() != "lola" {
		t.Fatalf("unexpected sources: %+v", srcs)
	}
}

func TestParseSourceNames_default(t *testing.T) {
	names := ParseSourceNames("")
	if len(names) != 1 || names[0] != "ds" {
		t.Fatalf("names = %v", names)
	}
}

func TestScrapeDeps_Publisher_missing(t *testing.T) {
	deps := &ScrapeDeps{Log: slog.Default(), Publishers: map[string]RawPublisher{}}
	_, err := deps.Publisher("ds")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockSource_policy(t *testing.T) {
	_ = ledger.PolicyStatic
}
