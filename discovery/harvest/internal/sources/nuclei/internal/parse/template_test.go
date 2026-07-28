package parse

import "testing"

func TestParseYAML_minimal(t *testing.T) {
	raw := []byte(`id: test-rule
info:
  name: Test Name
  severity: high
  tags: cve,http
  classification:
    cve-id: CVE-2023-12345
    cwe-id: CWE-79
`)
	p, err := ParseYAML(raw)
	if err != nil {
		t.Fatal(err)
	}
	if p.TemplateID != "test-rule" || p.Name != "Test Name" || p.Severity != "high" {
		t.Fatalf("unexpected: %+v", p)
	}
	if p.CVE != "CVE-2023-12345" || p.CWE != "CWE-79" {
		t.Fatalf("cve/cwe: %+v", p)
	}
}
