package usecase

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCalderaYAMLSequence(t *testing.T) {
	raw := `---
- id: test-ability-1
  name: Sample
  description: Desc
  tactic: collection
  technique:
    attack_id: T1005
    name: Data from Local System
`
	var seq []map[string]any
	if err := yaml.Unmarshal([]byte(raw), &seq); err != nil {
		t.Fatal(err)
	}
	if len(seq) != 1 || seq[0]["id"] != "test-ability-1" {
		t.Fatalf("%v", seq)
	}
}
