package retrieval

import "testing"

func TestApplyExactBoost_mitre(t *testing.T) {
	ranked := []RankedHit{
		{SkillID: "other", Score: 1.0, Hit: ChunkHit{Text: "generic"}},
		{SkillID: "powershell", Score: 0.5, Hit: ChunkHit{Text: "T1059.001 execution"}},
	}
	out := ApplyExactBoost("T1059.001", ranked)
	if out[0].SkillID != "powershell" {
		t.Fatalf("expected boost for T1059.001, top=%s", out[0].SkillID)
	}
}

func TestApplyExactBoost_noIdentifiers(t *testing.T) {
	ranked := []RankedHit{{SkillID: "a", Score: 1.0}}
	out := ApplyExactBoost("generic query", ranked)
	if out[0].SkillID != "a" {
		t.Fatal("expected unchanged order")
	}
}
