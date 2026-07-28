package usecase

import (
	"testing"
)

func TestParseYaraRuleName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		body     string
		fallback string
		want     string
	}{
		{
			name: "rule with brace",
			body: `// comment
rule EvilLoader {
  strings:
    $a = "x"
}`,
			fallback: "file.yar",
			want:     "EvilLoader",
		},
		{
			name:     "rule with tab before brace",
			body:     "rule Foo\t{\n",
			fallback: "file.yar",
			want:     "Foo",
		},
		{
			name:     "no rule keyword",
			body:     "import \"pe\"\n",
			fallback: "fallback.yar",
			want:     "fallback",
		},
		{
			name:     "empty body",
			body:     "",
			fallback: "x.yar",
			want:     "x",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := parseYaraRuleName(tt.body, tt.fallback); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestTagsToJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, "[]"},
		{"not array", "tags", "[]"},
		{"strings", []any{"a", "b"}, `["a","b"]`},
		{"skips non-strings", []any{"ok", 1, "z"}, `["ok","z"]`},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tagsToJSON(tt.in); got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestFirstNonEmpty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b string
		want string
	}{
		{"primary", "backup", "primary"},
		{"  ", "backup", "backup"},
		{"", "only", "only"},
	}
	for _, tt := range tests {
		if got := firstNonEmpty(tt.a, tt.b); got != tt.want {
			t.Fatalf("firstNonEmpty(%q,%q)=%q want %q", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestEnvInt(t *testing.T) {
	const key = "DS_TEST_ENV_INT"
	t.Setenv(key, "")
	if got := envInt(key, 99); got != 99 {
		t.Fatalf("empty: got %d", got)
	}
	t.Setenv(key, "7")
	if got := envInt(key, 99); got != 7 {
		t.Fatalf("valid: got %d", got)
	}
	t.Setenv(key, "bad")
	if got := envInt(key, 99); got != 99 {
		t.Fatalf("invalid: got %d", got)
	}
}
