package procedure

import (
	"testing"

	"github.com/butbeautifulv/veil/pkg/playbook/testutil"
)

func TestDefault_and_Meta(t *testing.T) {
	testutil.SetRepoRoot(t)
	cat, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	meta := cat.Meta()
	if len(meta.Procedures) == 0 {
		t.Fatal("expected procedures")
	}
	if _, ok := cat.GetSummary(meta.Procedures[0].ID); !ok {
		t.Fatal("GetSummary miss")
	}
}
