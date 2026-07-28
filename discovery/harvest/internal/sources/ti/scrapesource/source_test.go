package scrapesource

import "testing"

func TestSourceName(t *testing.T) {
	s := &Source{}
	if s.Name() != "ti" {
		t.Fatalf("Name() = %q", s.Name())
	}
}
