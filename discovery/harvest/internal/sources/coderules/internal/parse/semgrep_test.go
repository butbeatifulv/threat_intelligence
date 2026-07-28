package parse

import "testing"

func TestSemgrepCWES_metadataString(t *testing.T) {
	root := map[string]any{
		"rules": []any{
			map[string]any{
				"metadata": map[string]any{"cwe": "CWE-79: XSS"},
			},
		},
	}
	got := SemgrepCWES(root)
	if len(got) != 1 || got[0] != "CWE-79" {
		t.Fatalf("got %v", got)
	}
}
