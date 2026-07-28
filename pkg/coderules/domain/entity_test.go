package domain

import "testing"

func TestRuleFile_zeroSafe(t *testing.T) {
	var r RuleFile
	if r.Repo != "" || r.Format != "" {
		t.Fatal("zero value should be empty")
	}
}
