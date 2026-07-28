package domain

import "testing"

func TestAdvisoryRef_zeroSafe(t *testing.T) {
	var a AdvisoryRef
	if a.CVE != "" || a.Path != "" {
		t.Fatal("zero value should be empty")
	}
}
