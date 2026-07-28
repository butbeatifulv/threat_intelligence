package domain

import "testing"

func TestContourValid(t *testing.T) {
	for _, c := range []Contour{ContourIngest, ContourEngage, ContourKnowledge} {
		if !c.Valid() {
			t.Fatalf("%q should be valid", c)
		}
	}
	if Contour("pipeline").Valid() {
		t.Fatal("unknown contour should be invalid")
	}
}
