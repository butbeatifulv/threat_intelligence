package domain

import (
	"encoding/json"
	"testing"
)

func TestSkillMeta_JSONRoundTrip(t *testing.T) {
	in := SkillMeta{
		ID:          "skill-1",
		Name:        "Web Recon",
		Domain:      "offensive",
		Subdomain:   "recon",
		Description: "passive web recon",
		Tags:        []string{"web"},
		NISTCSF:     []string{"ID.RA"},
		Version:     "1.0",
		License:     "MIT",
		AttackIDs:   []string{"T1595"},
		CorpusPath:  "corpus/skills/web-recon",
		BodyChars:   1200,
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out SkillMeta
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.ID != in.ID || out.Name != in.Name || len(out.Tags) != 1 || out.BodyChars != in.BodyChars {
		t.Fatalf("got %+v", out)
	}
}

func TestProcedureSpec_JSONRoundTrip(t *testing.T) {
	in := ProcedureSpec{
		ID:        "proc-1",
		Subdomain: "recon",
		AttackIDs: []string{"T1595"},
		NISTCSF:   []string{"ID.RA"},
		WhenToUse: []string{"external assessment"},
		Prerequisites: []string{"scope approval"},
		Steps: []ProcedureStep{
			{
				Number:       1,
				Title:        "Enumerate",
				Kind:         StepShell,
				Body:         "run passive DNS",
				ToolMentions: []string{"amass"},
				CatalogTools: []string{"tool-amass"},
			},
		},
		ToolMentions: []string{"amass"},
		CatalogTools: []string{"tool-amass"},
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out ProcedureSpec
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.ID != in.ID || len(out.Steps) != 1 || out.Steps[0].Kind != StepShell {
		t.Fatalf("got %+v", out)
	}
}

func TestStepKind_constants(t *testing.T) {
	cases := []StepKind{StepShell, StepManual, StepTool}
	for _, c := range cases {
		if string(c) == "" {
			t.Fatalf("empty StepKind constant")
		}
	}
}

func TestSkillMeta_ProcedureSpec_zeroSafe(t *testing.T) {
	var m SkillMeta
	var p ProcedureSpec
	if m.ID != "" || m.Name != "" || len(m.Tags) != 0 {
		t.Fatal("zero SkillMeta should be empty")
	}
	if p.ID != "" || len(p.Steps) != 0 {
		t.Fatal("zero ProcedureSpec should be empty")
	}
}
