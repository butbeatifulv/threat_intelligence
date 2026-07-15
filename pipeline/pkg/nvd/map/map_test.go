package nvdmap_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/vuln/domain"
	"github.com/butbeautifulv/veil/pipeline/pkg/nvd/parse"
)

func TestFromNVD_mapsFixtureToCanonical(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "parse", "testdata", "nvd_page_min.json"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, total, err := parse.ParsePage(raw)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(parsed) != 2 {
		t.Fatalf("parsed: total=%d len=%d", total, len(parsed))
	}

	m := parsed[0]
	if m.CVE != "CVE-2024-0001" || m.ID != "CVE-2024-0001" {
		t.Fatalf("id/cve: %#v", m)
	}
	if m.Summary == "" {
		t.Fatal("expected summary")
	}
	if len(m.CWE) != 1 || m.CWE[0] != "CWE-79" {
		t.Fatalf("cwe: %#v", m.CWE)
	}
	if len(m.CPEs) != 1 || m.CPEs[0].URI == "" {
		t.Fatalf("cpes: %#v", m.CPEs)
	}
	if m.CVSS == nil || m.CVSS.Base != 9.8 || m.CVSS.Version != "3.1" {
		t.Fatalf("cvss: %#v", m.CVSS)
	}
}

func TestFromNVD_toCommitEnvelopes(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "parse", "testdata", "nvd_page_min.json"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, _, err := parse.ParsePage(raw)
	if err != nil {
		t.Fatal(err)
	}

	var envelopes []*commit.Envelope
	for _, p := range parsed {
		m := p
		e, err := commit.NewEnvelope(commit.SourceVuln, commit.KindVulnUpsert, commit.VulnUpsertIdempotencyKey(m.CVE), m)
		if err != nil {
			t.Fatal(err)
		}
		envelopes = append(envelopes, e)
	}

	if len(envelopes) != 2 {
		t.Fatalf("envelopes: got %d want 2", len(envelopes))
	}
	if envelopes[0].Kind != commit.KindVulnUpsert || envelopes[0].Source != commit.SourceVuln {
		t.Fatalf("first envelope: source=%s kind=%s", envelopes[0].Source, envelopes[0].Kind)
	}
	var v0 domain.Vulnerability
	if err := json.Unmarshal(envelopes[0].Payload, &v0); err != nil {
		t.Fatal(err)
	}
	if v0.CVE != "CVE-2024-0001" || len(v0.CWE) == 0 || len(v0.CPEs) == 0 || v0.CVSS == nil {
		t.Fatalf("payload: %#v", v0)
	}
	if envelopes[0].IdempotencyKey != commit.VulnUpsertIdempotencyKey("CVE-2024-0001") {
		t.Fatalf("idempotency %q", envelopes[0].IdempotencyKey)
	}
}
