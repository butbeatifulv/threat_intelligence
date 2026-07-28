package domain

import "testing"

func TestResource_zeroSafe(t *testing.T) {
	var r Resource
	if r.Key != "" || r.Source != "" {
		t.Fatal("zero value should be empty")
	}
}
