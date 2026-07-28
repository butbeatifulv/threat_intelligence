package procedure

import (
	"testing"

	"github.com/butbeautifulv/veil/pkg/playbook/domain"
)

func TestCatalogToolsForTechnique(t *testing.T) {
	procedures := []domain.ProcedureSummary{
		{
			ID:           "a",
			AttackIDs:    []string{"T1059.001"},
			CatalogTools: []string{"nmap_scan", "nuclei_scan"},
		},
		{
			ID:           "b",
			AttackIDs:    []string{"t1059.001"},
			CatalogTools: []string{"nmap_scan"},
		},
		{
			ID:           "c",
			AttackIDs:    []string{"T1005"},
			CatalogTools: []string{"ignored_tool"},
		},
	}
	got := CatalogToolsForTechnique("T1059.001", procedures)
	if len(got) != 2 {
		t.Fatalf("got %v want [nmap_scan nuclei_scan]", got)
	}
	if got[0] != "nmap_scan" || got[1] != "nuclei_scan" {
		t.Fatalf("order/dedup: %v", got)
	}
	if len(CatalogToolsForTechnique("T9999", procedures)) != 0 {
		t.Fatal("expected no tools for unknown technique")
	}
	if len(CatalogToolsForTechnique("  ", procedures)) != 0 {
		t.Fatal("expected no tools for blank technique")
	}
}
