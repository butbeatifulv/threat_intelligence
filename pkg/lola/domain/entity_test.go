package domain

import (
	"encoding/json"
	"testing"
)

func TestArtifact_JSONRoundTrip(t *testing.T) {
	in := Artifact{
		Name:        "cmd.exe",
		Description: "Windows command interpreter",
		OS:          []string{"windows"},
		Commands: []Command{
			{
				Command:     "cmd /c whoami",
				Description: "identity",
				Usecase:     "recon",
				Category:    "execution",
				Privileges:  "user",
				MitreID:     "T1059.003",
				OS:          []string{"windows"},
				Tags:        []string{"builtin"},
			},
		},
		Paths:     []string{"C:\\Windows\\System32\\cmd.exe"},
		Detection: Detection{Sigma: []string{"rule-1"}, Yara: []string{"rule-2"}},
		MitreID:   "T1059.003",
		Category:  "execution",
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out Artifact
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.Name != in.Name || len(out.Commands) != 1 || out.Commands[0].MitreID != in.Commands[0].MitreID {
		t.Fatalf("got %+v", out)
	}
	if len(out.Detection.Sigma) != 1 || out.Detection.Yara[0] != "rule-2" {
		t.Fatalf("detection %+v", out.Detection)
	}
}

func TestCommand_Detection_zeroSafe(t *testing.T) {
	var cmd Command
	var det Detection
	var art Artifact
	if cmd.Command != "" || cmd.MitreID != "" || len(cmd.OS) != 0 {
		t.Fatal("zero Command should be empty")
	}
	if len(det.Sigma) != 0 || len(det.Yara) != 0 {
		t.Fatal("zero Detection should be empty")
	}
	if art.Name != "" || len(art.Commands) != 0 {
		t.Fatal("zero Artifact should be empty")
	}
}
