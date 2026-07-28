package domain

import "testing"

func TestSourceRefZeroAndValid(t *testing.T) {
	var z SourceRef
	if !z.IsZero() {
		t.Fatal("empty ref should be zero")
	}
	if z.Valid() {
		t.Fatal("zero ref should be invalid")
	}
	r := SourceRef{Source: SourceTI, Key: "ioc-1"}
	if r.IsZero() || !r.Valid() {
		t.Fatalf("ti ref should be valid: %+v", r)
	}
	r2 := SourceRef{Source: Source("nope"), Key: "k"}
	if r2.Valid() {
		t.Fatal("invalid source should fail Valid")
	}
}
