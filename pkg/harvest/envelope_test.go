package harvest

import (
	"encoding/json"
	"testing"
)

func TestSourceConstants(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"SourceSBOM", SourceSBOM, "sbom"},
		{"SourceCoderules", SourceCoderules, "coderules"},
		{"SourceNuclei", SourceNuclei, "nuclei"},
		{"SourceTI", SourceTI, "ti"},
		{"SourceVuln", SourceVuln, "vuln"},
		{"SourceLola", SourceLola, "lola"},
		{"SourceDS", SourceDS, "ds"},
		{"SourceBrowser", SourceBrowser, "browser"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q want %q", tc.got, tc.want)
			}
			if tc.got == "" {
				t.Fatal("empty source constant")
			}
		})
	}
}

func TestKindConstants(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"KindSBOMOSVJSON", KindSBOMOSVJSON, "scrape_sbom_osv_json"},
		{"KindSBOMGHSAPath", KindSBOMGHSAPath, "scrape_sbom_ghsa_path"},
		{"KindCoderulesCWERaw", KindCoderulesCWERaw, "scrape_coderules_cwe_raw"},
		{"KindCoderulesSemgrepRaw", KindCoderulesSemgrepRaw, "scrape_coderules_semgrep_raw"},
		{"KindCoderulesCodeQLRaw", KindCoderulesCodeQLRaw, "scrape_coderules_codeql_raw"},
		{"KindNucleiTemplateRaw", KindNucleiTemplateRaw, "scrape_nuclei_template_raw"},
		{"KindTIKEVRow", KindTIKEVRow, "scrape_ti_kev_row"},
		{"KindTIJSONLLine", KindTIJSONLLine, "scrape_ti_jsonl_line"},
		{"KindTIIoCRaw", KindTIIoCRaw, "scrape_ti_ioc_raw"},
		{"KindTIReportRaw", KindTIReportRaw, "scrape_ti_report_raw"},
		{"KindTICampaignRaw", KindTICampaignRaw, "scrape_ti_campaign_raw"},
		{"KindTIClusterRaw", KindTIClusterRaw, "scrape_ti_cluster_raw"},
		{"KindTIActorRaw", KindTIActorRaw, "scrape_ti_actor_raw"},
		{"KindVulnCVEUpsert", KindVulnCVEUpsert, "scrape_vuln_cve_upsert"},
		{"KindVulnMergeExploit", KindVulnMergeExploit, "scrape_vuln_merge_exploit"},
		{"KindVulnNVDPage", KindVulnNVDPage, "scrape_nvd_page"},
		{"KindLolaArtifactRaw", KindLolaArtifactRaw, "scrape_lola_artifact_raw"},
		{"KindLolaLoftsRaw", KindLolaLoftsRaw, "scrape_lola_lofts_raw"},
		{"KindLolaLinkArtifacts", KindLolaLinkArtifacts, "scrape_lola_link_artifacts"},
		{"KindLolaAttackTechnique", KindLolaAttackTechnique, "scrape_lola_attack_technique"},
		{"KindLolaAttackTactic", KindLolaAttackTactic, "scrape_lola_attack_tactic"},
		{"KindLolaMergeTacticTechnique", KindLolaMergeTacticTechnique, "scrape_lola_merge_tactic_technique"},
		{"KindLolaMergeSubtechnique", KindLolaMergeSubtechnique, "scrape_lola_merge_subtechnique"},
		{"KindDSSigmaRaw", KindDSSigmaRaw, "scrape_ds_sigma_raw"},
		{"KindDSYaraRaw", KindDSYaraRaw, "scrape_ds_yara_raw"},
		{"KindDSAtomicRaw", KindDSAtomicRaw, "scrape_ds_atomic_raw"},
		{"KindDSCalderaRaw", KindDSCalderaRaw, "scrape_ds_caldera_raw"},
		{"KindBrowserInspectRaw", KindBrowserInspectRaw, "scrape_browser_inspect_raw"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q want %q", tc.got, tc.want)
			}
		})
	}
}

// domainRegistrySources mirrors pkg/domain/source.go allSources (wire SOT).
func domainRegistrySources() []string {
	return []string{
		"sbom", "coderules", "nuclei", "ti", "vuln", "lola", "ds", "browser", "engage",
	}
}

func TestSourcesAlignDomainRegistry(t *testing.T) {
	want := []string{
		SourceSBOM, SourceCoderules, SourceNuclei, SourceTI,
		SourceVuln, SourceLola, SourceDS, SourceBrowser,
	}
	registry := make(map[string]struct{})
	for _, s := range domainRegistrySources() {
		registry[s] = struct{}{}
	}
	for _, wire := range want {
		if _, ok := registry[wire]; !ok {
			t.Fatalf("domain registry missing harvest source %q", wire)
		}
	}
}

func TestNewEnvelope_validationErrors(t *testing.T) {
	t.Run("nil payload marshals null", func(t *testing.T) {
		var p *struct{}
		if _, err := NewEnvelope(SourceDS, KindDSSigmaRaw, "k", p); err == nil {
			t.Fatal("expected validation error for null payload")
		}
	})

	t.Run("unmarshalable payload", func(t *testing.T) {
		if _, err := NewEnvelope(SourceDS, KindDSSigmaRaw, "k", make(chan int)); err == nil {
			t.Fatal("expected marshal error")
		}
	})

	t.Run("empty content key", func(t *testing.T) {
		if _, err := NewEnvelope(SourceDS, KindDSSigmaRaw, "  ", DSSigmaRaw{Path: "x", RawYAML: "id: x"}); err == nil {
			t.Fatal("expected empty content_key error")
		}
	})
}

func TestEnvelopeValidate_edgeCases(t *testing.T) {
	validPayload := []byte(`{"x":1}`)
	base := Envelope{
		SchemaVersion: CurrentSchemaVersion,
		Source:        SourceDS,
		Kind:          KindDSSigmaRaw,
		ContentKey:    "ds:sigma:x",
		ScrapedAt:     "2026-05-20T00:00:00Z",
		Payload:       validPayload,
	}

	cases := []struct {
		name string
		mut  func(*Envelope)
	}{
		{"bad schema", func(e *Envelope) { e.SchemaVersion = 0 }},
		{"empty source", func(e *Envelope) { e.Source = "  " }},
		{"empty kind", func(e *Envelope) { e.Kind = "" }},
		{"empty content_key", func(e *Envelope) { e.ContentKey = "" }},
		{"empty scraped_at", func(e *Envelope) { e.ScrapedAt = "" }},
		{"empty payload", func(e *Envelope) { e.Payload = nil }},
		{"null payload", func(e *Envelope) { e.Payload = []byte("null") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := base
			tc.mut(&e)
			if err := e.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestDSContentKey(t *testing.T) {
	cases := []struct {
		prefix, path, want string
	}{
		{"sigma", "rules/x.yml", "ds:sigma:rules/x.yml"},
		{"yara", "  rules/y.yar  ", "ds:yara:rules/y.yar"},
		{"atomic", "", "ds:atomic:"},
	}
	for _, tc := range cases {
		t.Run(tc.prefix+"_"+tc.path, func(t *testing.T) {
			if got := DSContentKey(tc.prefix, tc.path); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestBrowserContentKey(t *testing.T) {
	cases := []struct {
		url, want string
	}{
		{"https://example.com", "browser:inspect:https://example.com"},
		{"  https://example.com/path  ", "browser:inspect:https://example.com/path"},
		{"", "browser:inspect:"},
	}
	for _, tc := range cases {
		t.Run(tc.url, func(t *testing.T) {
			if got := BrowserContentKey(tc.url); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestEnvelopeValidate(t *testing.T) {
	env, err := NewEnvelope(SourceDS, KindDSSigmaRaw, "ds:sigma:rules/x.yml", DSSigmaRaw{Path: "rules/x.yml", RawYAML: "id: x"})
	if err != nil {
		t.Fatal(err)
	}
	if err := env.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestEnvelopeBrowserInspect(t *testing.T) {
	env, err := NewEnvelope(SourceBrowser, KindBrowserInspectRaw, BrowserContentKey("https://example.com"), BrowserInspectRaw{
		URL: "https://example.com", RawJSON: `{"success":true}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := env.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestEnvelopeValidateEmptyPayload(t *testing.T) {
	env := &Envelope{SchemaVersion: 1, Source: SourceDS, Kind: KindDSSigmaRaw, ContentKey: "k"}
	if err := env.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestNewEnvelope_roundtripJSON(t *testing.T) {
	env, err := NewEnvelope(SourceTI, KindTIKEVRow, "ti:kev:CVE-2024-1", TIKEVRow{CVEID: "CVE-2024-1"})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	var out Envelope
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if err := out.Validate(); err != nil {
		t.Fatal(err)
	}
}
