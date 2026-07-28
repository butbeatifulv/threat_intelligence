package ds

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
)

func TestTransform_table(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name      string
		kind      string
		payload   any
		wantLen   int
		wantKind  string
		checkFirst func(t *testing.T, env *commit.Envelope)
	}{
		{
			name: "sigma",
			kind: harvest.KindDSSigmaRaw,
			payload: harvest.DSSigmaRaw{
				Path: "rules/test.yml",
				RawYAML: `id: test-sigma-1
title: Test Sigma
level: high
logsource:
  product: windows
  service: security
tags:
  - attack.execution
`,
			},
			wantLen:  1,
			wantKind: commit.KindDSUpsertSigma,
			checkFirst: func(t *testing.T, env *commit.Envelope) {
				var pl commit.DSUpsertSigmaPayload
				if err := json.Unmarshal(env.Payload, &pl); err != nil {
					t.Fatal(err)
				}
				if pl.ID != "test-sigma-1" || pl.Title != "Test Sigma" || pl.Level != "high" {
					t.Fatalf("payload: %#v", pl)
				}
				if pl.LogProduct != "windows" || pl.LogService != "security" {
					t.Fatalf("logsource: %#v", pl)
				}
			},
		},
		{
			name: "yara",
			kind: harvest.KindDSYaraRaw,
			payload: harvest.DSYaraRaw{
				Path: "rules/test.yar",
				RawBody: `rule TestRule {
  strings:
    $a = "test"
  condition:
    $a
}`,
			},
			wantLen:  1,
			wantKind: commit.KindDSUpsertYara,
			checkFirst: func(t *testing.T, env *commit.Envelope) {
				var pl commit.DSUpsertYaraPayload
				if err := json.Unmarshal(env.Payload, &pl); err != nil {
					t.Fatal(err)
				}
				if pl.Name != "TestRule" {
					t.Fatalf("name: %q", pl.Name)
				}
			},
		},
		{
			name: "atomic",
			kind: harvest.KindDSAtomicRaw,
			payload: harvest.DSAtomicRaw{
				Path: "atomics/T1003.yml",
				RawYAML: `attack_technique: T1003
atomic_tests:
  - name: Dump LSASS
    auto_generated_guid: guid-abc
    tactics:
      - credential-access
    executor:
      name: powershell
      command: Get-Process lsass
`,
			},
			wantLen:  1,
			wantKind: commit.KindDSUpsertAtomic,
			checkFirst: func(t *testing.T, env *commit.Envelope) {
				var pl commit.DSUpsertAtomicPayload
				if err := json.Unmarshal(env.Payload, &pl); err != nil {
					t.Fatal(err)
				}
				if pl.ID != "guid-abc" || pl.Technique != "T1003" || pl.ExecName != "powershell" {
					t.Fatalf("payload: %#v", pl)
				}
			},
		},
		{
			name: "caldera",
			kind: harvest.KindDSCalderaRaw,
			payload: harvest.DSCalderaRaw{
				Path:     "abilities/test.yml",
				FileName: "test.yml",
				RawBody: `id: ability-1
name: List files
description: Enumerate files
tactic: collection
technique:
  attack_id: T1005
`,
			},
			wantLen:  1,
			wantKind: commit.KindDSUpsertCaldera,
			checkFirst: func(t *testing.T, env *commit.Envelope) {
				var pl commit.DSUpsertCalderaPayload
				if err := json.Unmarshal(env.Payload, &pl); err != nil {
					t.Fatal(err)
				}
				if pl.ID != "ability-1" || pl.TechniqueID != "T1005" || pl.Tactic != "collection" {
					t.Fatalf("payload: %#v", pl)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatal(err)
			}
			env := &harvest.Envelope{Kind: tt.kind, Payload: payload}
			out, err := Transform(ctx, env)
			if err != nil {
				t.Fatal(err)
			}
			if len(out) != tt.wantLen {
				t.Fatalf("len(out) = %d want %d", len(out), tt.wantLen)
			}
			if out[0].Kind != tt.wantKind {
				t.Fatalf("kind = %s want %s", out[0].Kind, tt.wantKind)
			}
			if out[0].Source != commit.SourceDS {
				t.Fatalf("source = %s", out[0].Source)
			}
			if tt.checkFirst != nil {
				tt.checkFirst(t, out[0])
			}
		})
	}
}
