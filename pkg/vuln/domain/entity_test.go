package domain

import (
	"encoding/json"
	"testing"
)

func TestVulnerability_JSONRoundTrip(t *testing.T) {
	in := Vulnerability{
		ID:      "vuln-1",
		CVE:     "CVE-2024-0001",
		Summary: "test summary",
		CWE:     []string{"CWE-79"},
		CPEs:    []CPE{{URI: "cpe:2.3:a:vendor:product:1.0"}},
		CVSS:    &CVSS{Version: "3.1", Base: 9.8, Vector: "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"},
		Exploits: []ExploitRef{
			{Source: "exploit-db", RefID: "12345", URL: "https://example.com/e/1"},
		},
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out Vulnerability
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.CVE != in.CVE || out.CVSS == nil || out.CVSS.Base != in.CVSS.Base || len(out.Exploits) != 1 {
		t.Fatalf("got %+v", out)
	}
}

func TestCVSS_CPE_ExploitRef_zeroSafe(t *testing.T) {
	var c CVSS
	var cpe CPE
	var e ExploitRef
	var v Vulnerability
	if c.Version != "" || c.Vector != "" || c.Base != 0 {
		t.Fatal("zero CVSS should be empty")
	}
	if cpe.URI != "" {
		t.Fatal("zero CPE should be empty")
	}
	if e.Source != "" || e.RefID != "" || e.URL != "" {
		t.Fatal("zero ExploitRef should be empty")
	}
	if v.ID != "" || v.CVE != "" || v.CVSS != nil || len(v.CPEs) != 0 {
		t.Fatal("zero Vulnerability should be empty")
	}
}
