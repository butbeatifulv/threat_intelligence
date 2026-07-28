package domain

import "testing"

func TestTemplate_zeroSafe(t *testing.T) {
	var tpl Template
	if tpl.Repo != "" || tpl.ID != "" {
		t.Fatal("zero value should be empty")
	}
}
