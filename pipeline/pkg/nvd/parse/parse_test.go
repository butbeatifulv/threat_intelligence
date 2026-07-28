package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/butbeautifulv/veil/pipeline/pkg/nvd/parse"
)

func TestParsePage_CWEAndCPE(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "nvd_page_min.json"))
	if err != nil {
		t.Fatal(err)
	}
	vulns, total, err := parse.ParsePage(raw)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 {
		t.Fatalf("totalResults: got %d want 2", total)
	}
	if len(vulns) != 2 {
		t.Fatalf("vulns: got %d want 2", len(vulns))
	}
	v0 := vulns[0]
	if v0.CVE != "CVE-2024-0001" {
		t.Fatalf("cve: %q", v0.CVE)
	}
	if len(v0.CWE) != 1 || v0.CWE[0] != "CWE-79" {
		t.Fatalf("cwe: %#v", v0.CWE)
	}
	if len(v0.CPEs) != 1 || v0.CPEs[0].URI == "" {
		t.Fatalf("cpes: %#v", v0.CPEs)
	}
	if v0.CVSS == nil || v0.CVSS.Base != 9.8 {
		t.Fatalf("cvss: %#v", v0.CVSS)
	}
	v1 := vulns[1]
	if len(v1.CWE) != 1 || v1.CWE[0] != "CWE-89" {
		t.Fatalf("cwe v1: %#v", v1.CWE)
	}
	if len(v1.CPEs) != 1 {
		t.Fatalf("cpes v1: %#v", v1.CPEs)
	}
}
