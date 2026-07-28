package usecase

import (
	"testing"
)

func TestMitreExternalID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		m    map[string]any
		want string
	}{
		{
			name: "attack pattern",
			m: map[string]any{
				"external_references": []any{
					map[string]any{"source_name": "mitre-attack", "external_id": "T1059.001"},
				},
			},
			want: "T1059.001",
		},
		{
			name: "tactic",
			m: map[string]any{
				"external_references": []any{
					map[string]any{"source_name": "MITRE-ATTACK", "external_id": "TA0002"},
				},
			},
			want: "TA0002",
		},
		{
			name: "wrong source skipped",
			m: map[string]any{
				"external_references": []any{
					map[string]any{"source_name": "cve", "external_id": "CVE-1"},
					map[string]any{"source_name": "mitre-attack", "external_id": "T1003"},
				},
			},
			want: "T1003",
		},
		{
			name: "no refs",
			m:    map[string]any{},
			want: "",
		},
		{
			name: "empty external_id",
			m: map[string]any{
				"external_references": []any{
					map[string]any{"source_name": "mitre-attack", "external_id": ""},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := mitreExternalID(tt.m); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestEnvIntLola(t *testing.T) {
	const key = "LOLA_TEST_ENV_INT"
	t.Setenv(key, "")
	if got := envIntLola(key, 42); got != 42 {
		t.Fatalf("empty: got %d", got)
	}
	t.Setenv(key, "10")
	if got := envIntLola(key, 42); got != 10 {
		t.Fatalf("valid: got %d", got)
	}
	t.Setenv(key, "-1")
	if got := envIntLola(key, 42); got != 42 {
		t.Fatalf("negative: got %d", got)
	}
	t.Setenv(key, "nope")
	if got := envIntLola(key, 42); got != 42 {
		t.Fatalf("invalid: got %d", got)
	}
}
