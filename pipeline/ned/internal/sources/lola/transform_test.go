package lola

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
		wantKind  string
		check func(t *testing.T, env *commit.Envelope)
	}{
		{
			name: harvest.KindLolaArtifactRaw,
			kind: harvest.KindLolaArtifactRaw,
			payload: harvest.LolaArtifactRaw{
				Source:  "misp",
				Path:    "obj/1.json",
				RawBody: `{"name":"sample-malware","type":"malware"}`,
			},
			wantKind: commit.KindLolaArtifact,
			check: func(t *testing.T, env *commit.Envelope) {
				var pl commit.LolaArtifactPayload
				if err := json.Unmarshal(env.Payload, &pl); err != nil {
					t.Fatal(err)
				}
				if pl.Source != "misp" {
					t.Fatalf("source: %q", pl.Source)
				}
				wantKey := commit.LolaArtifactIdempotencyKey("misp", "sample-malware")
				if env.IdempotencyKey != wantKey {
					t.Fatalf("idempotency %q want %q", env.IdempotencyKey, wantKey)
				}
			},
		},
		{
			name: harvest.KindLolaLoftsRaw,
			kind: harvest.KindLolaLoftsRaw,
			payload: harvest.LolaLoftsRaw{
				Title:    "LOFTS entry",
				Category: "technique",
				LinkURL:  "https://example.com/lofts/1",
				Markdown: "# LOFTS",
			},
			wantKind: commit.KindLolaLofts,
			check: func(t *testing.T, env *commit.Envelope) {
				var pl commit.LolaLoftsPayload
				if err := json.Unmarshal(env.Payload, &pl); err != nil {
					t.Fatal(err)
				}
				if pl.LinkURL != "https://example.com/lofts/1" || pl.Title != "LOFTS entry" {
					t.Fatalf("payload: %#v", pl)
				}
			},
		},
		{
			name: harvest.KindLolaAttackTechnique,
			kind: harvest.KindLolaAttackTechnique,
			payload: harvest.LolaAttackTechnique{
				ID: "T1059", Name: "Command and Scripting Interpreter", Description: "desc",
			},
			wantKind: commit.KindLolaAttackTechnique,
			check: func(t *testing.T, env *commit.Envelope) {
				var pl commit.LolaAttackTechniquePayload
				if err := json.Unmarshal(env.Payload, &pl); err != nil {
					t.Fatal(err)
				}
				if pl.ID != "T1059" {
					t.Fatalf("id: %q", pl.ID)
				}
			},
		},
		{
			name: harvest.KindLolaAttackTactic,
			kind: harvest.KindLolaAttackTactic,
			payload: harvest.LolaAttackTactic{
				ID: "TA0002", Name: "Execution",
			},
			wantKind: commit.KindLolaAttackTactic,
			check: func(t *testing.T, env *commit.Envelope) {
				if env.IdempotencyKey != commit.LolaTacticIdempotencyKey("TA0002") {
					t.Fatalf("idempotency %q", env.IdempotencyKey)
				}
			},
		},
		{
			name: harvest.KindLolaMergeTacticTechnique,
			kind: harvest.KindLolaMergeTacticTechnique,
			payload: harvest.LolaMergeTacticTechnique{
				TacticID: "TA0002", TechniqueID: "T1059",
			},
			wantKind: commit.KindLolaMergeTacticTechnique,
			check: func(t *testing.T, env *commit.Envelope) {
				want := commit.LolaMergeTacticTechniqueIdempotencyKey("TA0002", "T1059")
				if env.IdempotencyKey != want {
					t.Fatalf("idempotency %q want %q", env.IdempotencyKey, want)
				}
			},
		},
		{
			name: harvest.KindLolaLinkArtifacts,
			kind: harvest.KindLolaLinkArtifacts,
			payload: struct{}{},
			wantKind: commit.KindLolaLinkArtifacts,
			check: func(t *testing.T, env *commit.Envelope) {
				if env.IdempotencyKey != commit.LolaLinkArtifactsIdempotencyKey() {
					t.Fatalf("idempotency %q", env.IdempotencyKey)
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
			if len(out) != 1 {
				t.Fatalf("len(out) = %d", len(out))
			}
			if out[0].Kind != tt.wantKind || out[0].Source != commit.SourceLola {
				t.Fatalf("envelope source=%s kind=%s", out[0].Source, out[0].Kind)
			}
			if tt.check != nil {
				tt.check(t, out[0])
			}
		})
	}
}
