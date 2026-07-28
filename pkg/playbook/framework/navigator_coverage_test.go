package framework

import (
	"testing"

	"github.com/butbeautifulv/veil/pkg/playbook/testutil"
)

func TestLoadNavigatorLayer_repo(t *testing.T) {
	testutil.SetRepoRoot(t)
	layer, err := LoadNavigatorLayer()
	if err != nil {
		t.Fatal(err)
	}
	if layer.Name == "" {
		t.Fatal("expected layer name")
	}
	if len(layer.Techniques) < 10 {
		t.Fatalf("techniques: %d", len(layer.Techniques))
	}
}

func TestNavigatorLayer_Summarize_unit(t *testing.T) {
	layer := &NavigatorLayer{
		Name:        "Test Layer",
		Domain:      "enterprise-attack",
		Versions:    map[string]string{"attack": "14"},
		Techniques:  []NavigatorTechnique{{TechniqueID: "T1003", Score: 3.5}},
	}
	scores := layer.TechniqueScores()
	if scores["T1003"] != 3.5 {
		t.Fatalf("score: %v", scores["T1003"])
	}
	sum := layer.Summarize()
	if sum.TechniqueCount != 1 || sum.LayerName != "Test Layer" || sum.AttackVersion != "14" {
		t.Fatalf("summary: %+v", sum)
	}
}

func TestCoverageSummary_fields(t *testing.T) {
	testutil.SetRepoRoot(t)
	layer, err := LoadNavigatorLayer()
	if err != nil {
		t.Fatal(err)
	}
	sum := layer.Summarize()
	if sum.LayerName != layer.Name {
		t.Fatalf("layer_name: %q", sum.LayerName)
	}
	if sum.TechniqueCount != len(layer.Techniques) {
		t.Fatalf("technique_count: %d vs %d", sum.TechniqueCount, len(layer.Techniques))
	}
	if sum.Domain != layer.Domain {
		t.Fatalf("domain: %q", sum.Domain)
	}
	if sum.AttackVersion == "" && layer.Versions != nil {
		if v := layer.Versions["attack"]; v != "" && sum.AttackVersion != v {
			t.Fatalf("attack_version: %q", sum.AttackVersion)
		}
	}
	scores := layer.TechniqueScores()
	if len(scores) == 0 {
		t.Fatal("expected technique scores")
	}
}
