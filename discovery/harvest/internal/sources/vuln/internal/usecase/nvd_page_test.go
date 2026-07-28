package usecase

import "testing"

func TestNVDPageStats(t *testing.T) {
	raw := []byte(`{"totalResults":42,"vulnerabilities":[{},{}]}`)
	total, count, err := nvdPageStats(raw)
	if err != nil {
		t.Fatal(err)
	}
	if total != 42 {
		t.Fatalf("totalResults: got %d want 42", total)
	}
	if count != 2 {
		t.Fatalf("itemCount: got %d want 2", count)
	}
}

func TestNVDPageStats_emptyPage(t *testing.T) {
	raw := []byte(`{"totalResults":0,"vulnerabilities":[]}`)
	_, count, err := nvdPageStats(raw)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("itemCount: got %d want 0", count)
	}
}
