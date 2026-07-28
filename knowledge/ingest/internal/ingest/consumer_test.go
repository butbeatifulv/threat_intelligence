package ingest

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/knowledge/ingest/internal/components"
	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/nats-io/nats.go"
)

func TestValidateEnvelopeSource_table(t *testing.T) {
	tiIOC, _ := json.Marshal(map[string]string{"type": "ip", "value": "203.0.113.1"})
	vulnBody, _ := json.Marshal(map[string]string{"cve": "CVE-2024-1", "id": "CVE-2024-1", "summary": "s"})
	engagePayload, _ := json.Marshal(commit.EngageToolRunPayload{
		Tool: "nmap", Target: "127.0.0.1", Subject: "t", Success: true, At: "2026-05-16T00:00:00Z",
	})
	engageFinding, _ := json.Marshal(commit.EngageFindingPayload{
		Tool: "nuclei", Target: "https://example.com", Title: "open port", Severity: "info",
	})

	tests := []struct {
		name    string
		env     commit.Envelope
		wantErr string // empty means success
	}{
		{
			name: "ti_ioc_ok",
			env: commit.Envelope{
				SchemaVersion:  commit.CurrentSchemaVersion,
				Source:         commit.SourceTI,
				Kind:           commit.KindTIIoC,
				IdempotencyKey: commit.TIIoCIdempotencyKey("abc"),
				Payload:        tiIOC,
			},
		},
		{
			name: "ti_ioc_wrong_source",
			env: commit.Envelope{
				SchemaVersion:  commit.CurrentSchemaVersion,
				Source:         commit.SourceVuln,
				Kind:           commit.KindTIIoC,
				IdempotencyKey: commit.TIIoCIdempotencyKey("abc"),
				Payload:        tiIOC,
			},
			wantErr: `kind "ti_ioc" expects source "ti"`,
		},
		{
			name: "vuln_upsert_ok",
			env: commit.Envelope{
				SchemaVersion:  commit.CurrentSchemaVersion,
				Source:         commit.SourceVuln,
				Kind:           commit.KindVulnUpsert,
				IdempotencyKey: commit.VulnUpsertIdempotencyKey("CVE-2024-1"),
				Payload:        vulnBody,
			},
		},
		{
			name: "vuln_upsert_wrong_source",
			env: commit.Envelope{
				SchemaVersion:  commit.CurrentSchemaVersion,
				Source:         commit.SourceEngage,
				Kind:           commit.KindVulnUpsert,
				IdempotencyKey: commit.VulnUpsertIdempotencyKey("CVE-2024-1"),
				Payload:        vulnBody,
			},
			wantErr: `kind "vuln_upsert" expects source "vuln"`,
		},
		{
			name: "engage_tool_run_ok",
			env: commit.Envelope{
				SchemaVersion:  commit.CurrentSchemaVersion,
				Source:         commit.SourceEngage,
				Kind:           commit.KindEngageToolRun,
				IdempotencyKey: commit.EngageToolRunIdempotencyKey("nmap", "127.0.0.1", "2026-05-16T00:00:00Z"),
				Payload:        engagePayload,
			},
		},
		{
			name: "engage_finding_ok",
			env: commit.Envelope{
				SchemaVersion:  commit.CurrentSchemaVersion,
				Source:         commit.SourceEngage,
				Kind:           commit.KindEngageFinding,
				IdempotencyKey: commit.EngageFindingIdempotencyKey("nuclei", "https://example.com", "open port"),
				Payload:        engageFinding,
			},
		},
		{
			name: "engage_finding_wrong_source",
			env: commit.Envelope{
				SchemaVersion:  commit.CurrentSchemaVersion,
				Source:         commit.SourceTI,
				Kind:           commit.KindEngageFinding,
				IdempotencyKey: commit.EngageFindingIdempotencyKey("nuclei", "https://example.com", "open port"),
				Payload:        engageFinding,
			},
			wantErr: `kind "engage_finding" expects source "engage"`,
		},
		{
			name: "source_kind_mismatch",
			env: commit.Envelope{
				SchemaVersion:  commit.CurrentSchemaVersion,
				Source:         commit.SourceVuln,
				Kind:           commit.KindEngageToolRun,
				IdempotencyKey: "x",
				Payload:        engagePayload,
			},
			wantErr: `kind "engage_tool_run" expects source "engage"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateEnvelopeSource(&tc.env)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err %q want substring %q", err, tc.wantErr)
			}
		})
	}
}

func TestEnvelopeParse_validateAndSourceCheck(t *testing.T) {
	build := func(t *testing.T, source, kind string, payload any, keyFn func() string) commit.Envelope {
		t.Helper()
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		env := commit.Envelope{
			SchemaVersion:  commit.CurrentSchemaVersion,
			Source:         source,
			Kind:           kind,
			IdempotencyKey: keyFn(),
			Payload:        payloadBytes,
		}
		b, err := json.Marshal(env)
		if err != nil {
			t.Fatal(err)
		}
		var decoded commit.Envelope
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatal(err)
		}
		if err := decoded.Validate(); err != nil {
			t.Fatal(err)
		}
		if err := validateEnvelopeSource(&decoded); err != nil {
			t.Fatal(err)
		}
		return decoded
	}

	t.Run("ti", func(t *testing.T) {
		env := build(t, commit.SourceTI, commit.KindTIIoC,
			map[string]string{"type": "domain", "value": "evil.example"},
			func() string { return commit.TIIoCIdempotencyKey("node-1") },
		)
		if env.Source != commit.SourceTI || env.Kind != commit.KindTIIoC {
			t.Fatalf("%+v", env)
		}
	})

	t.Run("vuln", func(t *testing.T) {
		env := build(t, commit.SourceVuln, commit.KindVulnUpsert,
			map[string]string{"cve": "CVE-2025-100", "id": "CVE-2025-100", "summary": "test"},
			func() string { return commit.VulnUpsertIdempotencyKey("CVE-2025-100") },
		)
		if env.Source != commit.SourceVuln {
			t.Fatal(env.Source)
		}
	})

	t.Run("engage", func(t *testing.T) {
		env := build(t, commit.SourceEngage, commit.KindEngageToolRun,
			commit.EngageToolRunPayload{
				Tool: "nmap", Target: "10.0.0.1", Subject: "scan", Success: true, At: "2026-05-16T12:00:00Z",
			},
			func() string { return commit.EngageToolRunIdempotencyKey("nmap", "10.0.0.1", "2026-05-16T12:00:00Z") },
		)
		var p commit.EngageToolRunPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			t.Fatal(err)
		}
		if p.Tool != "nmap" || p.Target != "10.0.0.1" {
			t.Fatalf("%+v", p)
		}
	})

	t.Run("sbom_osv", func(t *testing.T) {
		env := build(t, commit.SourceSBOM, commit.KindSBOMOSVRecord,
			commit.SBOMOSVPayload{OSVID: "OSV-1", CVE: "CVE-2024-1", Affected: nil},
			func() string { return commit.SBOMOSVIdempotencyKey("CVE-2024-1", "osv", "OSV-1") },
		)
		if env.Source != commit.SourceSBOM || env.Kind != commit.KindSBOMOSVRecord {
			t.Fatalf("%+v", env)
		}
	})

	t.Run("coderules_cwe", func(t *testing.T) {
		env := build(t, commit.SourceCoderules, commit.KindCoderulesCWERow,
			commit.CoderulesCWEPayload{ID: "CWE-79", Name: "XSS"},
			func() string { return commit.CoderulesCWEIdempotencyKey("CWE-79") },
		)
		if env.Source != commit.SourceCoderules {
			t.Fatal(env.Source)
		}
	})
}

func TestHandleMsg_routesEngageToApplier(t *testing.T) {
	var called bool
	rt := &components.Runtime{
		Apply: components.DomainAppliers{
			Engage: func(_ context.Context, e *commit.Envelope) error {
				called = true
				if e.Kind != commit.KindEngageToolRun {
					t.Fatalf("kind %s", e.Kind)
				}
				return nil
			},
		},
	}
	payload, _ := json.Marshal(commit.EngageToolRunPayload{
		Tool: "nmap", Target: "127.0.0.1", Subject: "t", Success: true, At: "2026-05-16T00:00:00Z",
	})
	env := commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceEngage,
		Kind:           commit.KindEngageToolRun,
		IdempotencyKey: commit.EngageToolRunIdempotencyKey("nmap", "127.0.0.1", "2026-05-16T00:00:00Z"),
		Payload:        payload,
	}
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	if err := handleMsg(context.Background(), slog.Default(), &nats.Msg{Data: b}, rt); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("Engage applier not called")
	}
}

func TestHandleMsg_routesTIToApplier(t *testing.T) {
	var called bool
	rt := &components.Runtime{
		Apply: components.DomainAppliers{
			TI: func(_ context.Context, e *commit.Envelope) error {
				called = true
				if e.Kind != commit.KindTIIoC {
					t.Fatalf("kind %s", e.Kind)
				}
				return nil
			},
		},
	}
	payload, _ := json.Marshal(map[string]string{"type": "ip", "value": "198.51.100.1"})
	env := commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceTI,
		Kind:           commit.KindTIIoC,
		IdempotencyKey: commit.TIIoCIdempotencyKey("ioc-1"),
		Payload:        payload,
	}
	b, _ := json.Marshal(env)
	if err := handleMsg(context.Background(), slog.Default(), &nats.Msg{Data: b}, rt); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("TI applier not called")
	}
}

func TestHandleMsg_routesVulnToApplier(t *testing.T) {
	var called bool
	rt := &components.Runtime{
		Apply: components.DomainAppliers{
			Vuln: func(_ context.Context, e *commit.Envelope) error {
				called = true
				if e.Kind != commit.KindVulnUpsert {
					t.Fatalf("kind %s", e.Kind)
				}
				return nil
			},
		},
	}
	payload, _ := json.Marshal(map[string]string{"cve": "CVE-2024-42", "id": "CVE-2024-42", "summary": "x"})
	env := commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceVuln,
		Kind:           commit.KindVulnUpsert,
		IdempotencyKey: commit.VulnUpsertIdempotencyKey("CVE-2024-42"),
		Payload:        payload,
	}
	b, _ := json.Marshal(env)
	if err := handleMsg(context.Background(), slog.Default(), &nats.Msg{Data: b}, rt); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("Vuln applier not called")
	}
}
