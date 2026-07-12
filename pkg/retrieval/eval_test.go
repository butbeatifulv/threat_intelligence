package retrieval_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/butbeautifulv/veil/pkg/playbook/index"
	"github.com/butbeautifulv/veil/pkg/retrieval"
)

func TestEvalBaselineLegacy(t *testing.T) {
	root, err := retrieval.VeilRoot()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("VEIL_REPO_ROOT", root)
	evalPath := filepath.Join(root, retrieval.DefaultEvalRel)
	if _, err := os.Stat(evalPath); err != nil {
		t.Fatalf("missing %s", evalPath)
	}
	cat, err := index.Open("")
	if err != nil {
		t.Skip(err)
	}
	file, err := retrieval.LoadEvalFile(root, retrieval.DefaultEvalRel)
	if err != nil {
		t.Fatal(err)
	}
	m := retrieval.EvalLegacy(cat, file, 5)
	if m.Scored == 0 {
		t.Fatal("expected scored queries")
	}
	t.Logf("baseline recall@5=%.3f mrr=%.3f scored=%d/%d", m.RecallAt5, m.MRR, m.Scored, m.Total)
	for _, r := range m.PerQuery {
		if r.Skipped {
			continue
		}
		if !r.HitAt5 {
			t.Logf("miss: %s query=%q top=%s", r.ID, r.Query, r.TopID)
		}
	}
}

func TestEvalBleveEngine(t *testing.T) {
	root, err := retrieval.VeilRoot()
	if err != nil {
		t.Fatal(err)
	}
	cfg := retrieval.ConfigFromEnv(root)
	if _, err := os.Stat(cfg.BleveIndexPath); err != nil {
		t.Skip("bleve index missing; run make search-index")
	}
	eng, err := retrieval.NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer eng.Close()
	file, err := retrieval.LoadEvalFile(root, retrieval.DefaultEvalRel)
	if err != nil {
		t.Fatal(err)
	}
	m := retrieval.Eval(file, func(query, subdomain string, lim int) []string {
		ids, err := eng.SearchIDs(query, subdomain, lim)
		if err != nil {
			return nil
		}
		return ids
	}, 5)
	t.Logf("bleve recall@5=%.3f mrr=%.3f scored=%d/%d", m.RecallAt5, m.MRR, m.Scored, m.Total)
	if m.RecallAt5 < 0.5 {
		t.Fatalf("bleve recall@5 too low: %.3f", m.RecallAt5)
	}
}

func TestLoadEvalFile(t *testing.T) {
	root, err := retrieval.VeilRoot()
	if err != nil {
		t.Fatal(err)
	}
	f, err := retrieval.LoadEvalFile(root, retrieval.DefaultEvalRel)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Queries) < 20 {
		t.Fatalf("expected >=20 queries, got %d", len(f.Queries))
	}
}
