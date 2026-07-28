package query

import "testing"

func TestCategories_includesPlaybook(t *testing.T) {
	if !ValidCategory("playbook") {
		t.Fatal("expected playbook category")
	}
	meta, ok := Categories["playbook"]
	if !ok || meta.Title == "" {
		t.Fatalf("meta: %#v", meta)
	}
}

func TestCategories_includesEngage(t *testing.T) {
	if !ValidCategory("engage") {
		t.Fatal("expected engage category")
	}
	meta, ok := Categories["engage"]
	if !ok {
		t.Fatal("missing engage meta")
	}
	if len(meta.Labels) < 3 {
		t.Fatalf("labels: %v", meta.Labels)
	}
	ids := CategoryIDs()
	if ids[len(ids)-1] != "engage" {
		t.Fatalf("order: %v", ids)
	}
}
