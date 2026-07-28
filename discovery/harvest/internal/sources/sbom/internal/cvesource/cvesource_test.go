package cvesource

import (
	"context"
	"testing"
)

func TestParseCVEList(t *testing.T) {
	raw := `# comment
CVE-2024-0001

CVE-2023-9999
not-a-cve
CVE-2022-0002
`
	got, err := parseCVEList(context.Background(), raw, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "CVE-2024-0001" || got[1] != "CVE-2023-9999" {
		t.Fatalf("got %#v", got)
	}
}

func TestParseCVEListEmpty(t *testing.T) {
	_, err := parseCVEList(context.Background(), "# only\n", 10)
	if err == nil {
		t.Fatal("want error")
	}
}
