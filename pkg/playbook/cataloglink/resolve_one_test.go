package cataloglink

import "testing"

func TestResolveOne_aliasNotInCatalog(t *testing.T) {
	ResetCatalogForTest()
	catalogNames = map[string]struct{}{"other_tool": {}}
	if got := resolveOne("nmap"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveOne_prefixAndScanCandidate(t *testing.T) {
	ResetCatalogForTest()
	catalogNames = map[string]struct{}{
		"custom_scan": {},
		"xyz_tool":    {},
		"tool_exact":  {},
		"nmap_scan":   {},
	}
	if got := resolveOne("nmap"); got != "nmap_scan" {
		t.Fatalf("alias hit got %q", got)
	}
	if got := resolveOne("custom"); got != "custom_scan" {
		t.Fatalf("prefix match on custom_scan got %q", got)
	}
	if got := resolveOne("xyz"); got != "xyz_tool" {
		t.Fatalf("prefix got %q", got)
	}
	catalogNames["alpha_beta"] = struct{}{}
	if got := resolveOne("beta"); got != "alpha_beta" {
		t.Fatalf("suffix got %q", got)
	}
	if got := resolveOne("tool_exact"); got != "tool_exact" {
		t.Fatalf("exact got %q", got)
	}
	if got := resolveOne("ab"); got != "" {
		t.Fatalf("short token got %q", got)
	}
}

func TestResolveMentions_duplicateHits(t *testing.T) {
	ResetCatalogForTest()
	catalogOnce.Do(func() {})
	catalogNames = map[string]struct{}{"nmap_scan": {}}
	got := ResolveMentions([]string{"nmap", "nmap_scan"})
	if len(got) != 1 || got[0] != "nmap_scan" {
		t.Fatalf("got %v", got)
	}
}
