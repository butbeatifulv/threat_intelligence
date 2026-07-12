package retrieval

import "testing"

func TestRRF_twoLists(t *testing.T) {
	listA := []RankedHit{
		{SkillID: "a", Score: 1},
		{SkillID: "b", Score: 0.9},
		{SkillID: "c", Score: 0.8},
	}
	listB := []RankedHit{
		{SkillID: "b", Score: 1},
		{SkillID: "d", Score: 0.9},
		{SkillID: "a", Score: 0.8},
	}
	out := RRF([][]RankedHit{listA, listB}, 60, nil)
	if len(out) < 2 {
		t.Fatalf("expected fused hits, got %d", len(out))
	}
	if out[0].SkillID != "b" && out[0].SkillID != "a" {
		t.Fatalf("unexpected top: %s", out[0].SkillID)
	}
}
