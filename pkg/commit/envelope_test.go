package commit

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
		{"SourceEngage", SourceEngage, "engage"},
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

// domainRegistrySources mirrors pkg/domain/source.go allSources (wire SOT).
func domainRegistrySources() []string {
	return []string{
		"sbom", "coderules", "nuclei", "ti", "vuln", "lola", "ds", "browser", "engage",
	}
}

func TestSourcesAlignDomainRegistry(t *testing.T) {
	want := []string{
		SourceSBOM, SourceCoderules, SourceNuclei, SourceTI,
		SourceVuln, SourceLola, SourceDS, SourceEngage,
	}
	registry := make(map[string]struct{})
	for _, s := range domainRegistrySources() {
		registry[s] = struct{}{}
	}
	for _, wire := range want {
		if _, ok := registry[wire]; !ok {
			t.Fatalf("domain registry missing commit source %q", wire)
		}
	}
}

func TestAllIdempotencyKeys(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"SBOMOSV", SBOMOSVIdempotencyKey("  cve-2024-1 ", " PyPI ", " foo "), "sbom:osv:CVE-2024-1:pypi:foo"},
		{"SBOMGHSA", SBOMGHSAIdempotencyKey(" advisories/x.json "), "sbom:ghsa:advisories/x.json"},
		{"CoderulesCWE", CoderulesCWEIdempotencyKey(" cwe-79 "), "coderules:cwe:CWE-79"},
		{"CoderulesSemgrep", CoderulesSemgrepIdempotencyKey(" rules/x.yml "), "coderules:semgrep:rules/x.yml"},
		{"CoderulesCodeQL", CoderulesCodeQLIdempotencyKey(" queries/x.ql "), "coderules:codeql:queries/x.ql"},
		{"NucleiTemplate", NucleiTemplateIdempotencyKey(" http/cves/CVE-1.yaml "), "nuclei:template:http/cves/CVE-1.yaml"},
		{"TIIoC", TIIoCIdempotencyKey(" deadbeef "), "ti:ioc:deadbeef"},
		{"TIKEV", TIKEVIdempotencyKey(" cve-2024-0001 "), "ti:kev:CVE-2024-0001"},
		{"TIReport", TIReportIdempotencyKey(" report-stable-1 "), "ti:report:report-stable-1"},
		{"TICampaign", TICampaignIdempotencyKey(" camp-1 "), "ti:campaign:camp-1"},
		{"TICluster", TIClusterIdempotencyKey(" cluster-1 "), "ti:cluster:cluster-1"},
		{"TIActor", TIActorIdempotencyKey(" actor-hash "), "ti:actor:actor-hash"},
		{"TILinkCampaignIOC", TILinkCampaignIOCIdempotencyKey(" c1 ", " ioc1 "), "ti:lc:c1:ioc1"},
		{"TILinkClusterCampaign", TILinkClusterCampaignIdempotencyKey(" cl1 ", " c1 "), "ti:lkc:cl1:c1"},
		{"TILinkCampaignActor", TILinkCampaignActorIdempotencyKey(" c1 ", " a1 "), "ti:lca:c1:a1"},
		{"TILinkReportMentionsIOC", TILinkReportMentionsIOCIdempotencyKey(" r1 ", " ioc1 "), "ti:lrmi:r1:ioc1"},
		{"TIJSONLRecord", TIJSONLRecordIdempotencyKey(" abcdef0123456789abcdef0123456789 "), "ti:jsonl:abcdef0123456789abcdef0123456789"},
		{"VulnUpsert", VulnUpsertIdempotencyKey(" cve-2024-2 "), "vuln:upsert:CVE-2024-2"},
		{"VulnMergeExploit", VulnMergeExploitIdempotencyKey(" cve-2024-2 ", " epss ", " ref-9 "), "vuln:exp:CVE-2024-2:epss:ref-9"},
		{"LolaArtifact", LolaArtifactIdempotencyKey(" LOLBAS ", " cmd.exe "), "lola:art:lolbas:cmd.exe"},
		{"LolaLofts", LolaLoftsIdempotencyKey(" https://lofts.example/x "), "lola:lofts:https://lofts.example/x"},
		{"LolaTechnique", LolaTechniqueIdempotencyKey(" T1059 "), "lola:tech:T1059"},
		{"LolaTactic", LolaTacticIdempotencyKey(" TA0002 "), "lola:tac:TA0002"},
		{"LolaMergeTacticTechnique", LolaMergeTacticTechniqueIdempotencyKey(" TA0002 ", " T1059 "), "lola:mt:TA0002:T1059"},
		{"LolaMergeSubtechnique", LolaMergeSubtechniqueIdempotencyKey(" T1059 ", " T1059.001 "), "lola:ms:T1059:T1059.001"},
		{"LolaLinkArtifacts", LolaLinkArtifactsIdempotencyKey(), "lola:link:artifacts:v1"},
		{"DSSigma", DSSigmaIdempotencyKey(" sigma-uuid "), "ds:sigma:sigma-uuid"},
		{"DSYara", DSYaraIdempotencyKey(" yara-uuid "), "ds:yara:yara-uuid"},
		{"DSAtomic", DSAtomicIdempotencyKey(" atomic-uuid "), "ds:atomic:atomic-uuid"},
		{"DSCaldera", DSCalderaIdempotencyKey(" caldera-uuid "), "ds:caldera:caldera-uuid"},
		{"EngageToolRun", EngageToolRunIdempotencyKey(" nuclei ", " https://example.com ", " 2026-05-16T00:00:00Z "), "engage:run:nuclei:https://example.com:2026-05-16T00:00:00Z"},
		{"EngageFinding", EngageFindingIdempotencyKey(" nuclei ", " https://example.com ", " XSS "), "engage:finding:nuclei:https://example.com:XSS"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q want %q", tc.got, tc.want)
			}
			// stability: same inputs produce same key
			switch tc.name {
			case "SBOMOSV":
				if SBOMOSVIdempotencyKey("  cve-2024-1 ", " PyPI ", " foo ") != tc.want {
					t.Fatal("unstable")
				}
			case "LolaLinkArtifacts":
				if LolaLinkArtifactsIdempotencyKey() != tc.want {
					t.Fatal("unstable")
				}
			}
		})
	}
}

func TestEnvelope_Validate(t *testing.T) {
	p, err := NewEnvelope(SourceSBOM, KindSBOMOSVRecord, "k", SBOMOSVPayload{CVE: "CVE-2024-1", OSVID: "CVE-2024-1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	bad := &Envelope{SchemaVersion: 0, Source: SourceSBOM, Kind: KindSBOMOSVRecord, IdempotencyKey: "x", Payload: []byte(`{}`)}
	if bad.Validate() == nil {
		t.Fatal("expected error")
	}
}

func TestEnvelope_Validate_edgeCases(t *testing.T) {
	validPayload := []byte(`{"cve":"CVE-1"}`)
	base := Envelope{
		SchemaVersion:  CurrentSchemaVersion,
		Source:         SourceSBOM,
		Kind:           KindSBOMOSVRecord,
		IdempotencyKey: "sbom:osv:CVE-1:pypi:foo",
		Payload:        validPayload,
	}

	cases := []struct {
		name string
		mut  func(*Envelope)
	}{
		{"unsupported schema", func(e *Envelope) { e.SchemaVersion = 2 }},
		{"empty source", func(e *Envelope) { e.Source = "\t" }},
		{"empty kind", func(e *Envelope) { e.Kind = " " }},
		{"empty idempotency_key", func(e *Envelope) { e.IdempotencyKey = "" }},
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

func TestNewEnvelope_roundtrip(t *testing.T) {
	in := SBOMOSVPayload{CVE: "CVE-2024-0001", OSVID: "CVE-2024-0001", Affected: []map[string]any{{"package": map[string]any{"ecosystem": "PyPI", "name": "foo"}}}}
	e, err := NewEnvelope(SourceSBOM, KindSBOMOSVRecord, SBOMOSVIdempotencyKey("CVE-2024-0001", "PyPI", "foo"), in)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(e)
	var out Envelope
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if err := out.Validate(); err != nil {
		t.Fatal(err)
	}
	var p SBOMOSVPayload
	if err := json.Unmarshal(out.Payload, &p); err != nil {
		t.Fatal(err)
	}
	if p.CVE != in.CVE || len(p.Affected) != 1 {
		t.Fatalf("payload mismatch %+v", p)
	}
}

func TestNewEnvelope_marshalError(t *testing.T) {
	ch := make(chan int)
	_, err := NewEnvelope(SourceTI, KindTIIoC, "ti:ioc:1", ch)
	if err == nil {
		t.Fatal("expected json marshal error")
	}
}

func TestNewEnvelope_doesNotValidate(t *testing.T) {
	// NewEnvelope does not call Validate; callers must validate before publish.
	e, err := NewEnvelope(SourceSBOM, KindSBOMOSVRecord, "  ", SBOMOSVPayload{CVE: "CVE-1", OSVID: "CVE-1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := e.Validate(); err == nil {
		t.Fatal("expected empty idempotency_key error on Validate")
	}
}
