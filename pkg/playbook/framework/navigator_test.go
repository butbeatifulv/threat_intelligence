package framework

import (
	"testing"
)

func TestLoadNavigatorLayer(t *testing.T) {
	layer, err := LoadNavigatorLayer()
	if err != nil {
		t.Skip(err)
	}
	if len(layer.Techniques) < 100 {
		t.Fatalf("expected many techniques, got %d", len(layer.Techniques))
	}
	scores := layer.TechniqueScores()
	if scores["T1059.001"] <= 0 {
		t.Fatalf("expected T1059.001 score, got %v", scores["T1059.001"])
	}
	sum := layer.Summarize()
	if sum.TechniqueCount != len(layer.Techniques) {
		t.Fatalf("summary count %d vs %d", sum.TechniqueCount, len(layer.Techniques))
	}
}
