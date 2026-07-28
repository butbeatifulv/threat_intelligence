package domain

import (
	"testing"

	coderulesdomain "github.com/butbeautifulv/veil/pkg/coderules/domain"
	dsdomain "github.com/butbeautifulv/veil/pkg/ds/domain"
	nucleidomain "github.com/butbeautifulv/veil/pkg/nuclei/domain"
	sbomdomain "github.com/butbeautifulv/veil/pkg/sbom/domain"
)

func TestSourceRefFromAdvisoryRef(t *testing.T) {
	r := SourceRefFromAdvisoryRef(sbomdomain.AdvisoryRef{
		CVE: "CVE-2024-1", Source: "sbom", Path: "/data/osv.json",
	})
	if r.Source != SourceSBOM || r.Key != "CVE-2024-1" || r.Path != "/data/osv.json" || !r.Valid() {
		t.Fatalf("advisory ref: %+v", r)
	}
	r2 := SourceRefFromAdvisoryRef(sbomdomain.AdvisoryRef{CVE: "CVE-x", Path: "p"})
	if r2.Source != SourceSBOM {
		t.Fatalf("default source sbom: %+v", r2)
	}
}

func TestSourceRefFromTemplate(t *testing.T) {
	r := SourceRefFromTemplate(nucleidomain.Template{ID: "tpl-1", Path: "http/cves.yaml"})
	if r.Source != SourceNuclei || r.Key != "tpl-1" || !r.Valid() {
		t.Fatalf("template ref: %+v", r)
	}
	r2 := SourceRefFromTemplate(nucleidomain.Template{Path: "only-path.yaml"})
	if r2.Key != "only-path.yaml" {
		t.Fatalf("key falls back to path: %+v", r2)
	}
}

func TestSourceRefFromResource(t *testing.T) {
	r := SourceRefFromResource(dsdomain.Resource{Key: "sigma-1", Source: "ds", URL: "https://x", Kind: "sigma"})
	if r.Source != SourceDS || r.Key != "sigma-1" || r.Kind != "sigma" || !r.Valid() {
		t.Fatalf("resource ref: %+v", r)
	}
}

func TestSourceRefFromRuleFile(t *testing.T) {
	r := SourceRefFromRuleFile(coderulesdomain.RuleFile{Repo: "org/rules", Path: "cwe/CWE-79.yml", Format: "semgrep"})
	if r.Source != SourceCoderules || r.Key != "org/rules/cwe/CWE-79.yml" || !r.Valid() {
		t.Fatalf("rule file ref: %+v", r)
	}
	rRepo := SourceRefFromRuleFile(coderulesdomain.RuleFile{Repo: "solo-repo"})
	if rRepo.Key != "solo-repo" {
		t.Fatalf("repo-only key: %+v", rRepo)
	}
	rPath := SourceRefFromRuleFile(coderulesdomain.RuleFile{Path: "only.yml"})
	if rPath.Key != "only.yml" {
		t.Fatalf("path-only key: %+v", rPath)
	}
}

func TestSourceRefRoundTripFields(t *testing.T) {
	refs := []SourceRef{
		SourceRefFromAdvisoryRef(sbomdomain.AdvisoryRef{CVE: "CVE-1", Path: "p"}),
		SourceRefFromTemplate(nucleidomain.Template{ID: "id"}),
		SourceRefFromResource(dsdomain.Resource{Key: "k", URL: "u"}),
		SourceRefFromRuleFile(coderulesdomain.RuleFile{Path: "rules/x.yml"}),
	}
	for _, r := range refs {
		if r.IsZero() || !r.Valid() {
			t.Fatalf("adapter produced invalid ref: %+v", r)
		}
	}
}
