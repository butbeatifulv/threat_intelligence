package usecase

import (
	"context"
	"errors"
	"testing"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"github.com/butbeautifulv/veil/knowledge/connector/query"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/domain"
)

type fakeReadExec struct {
	out any
	err error
}

func (f *fakeReadExec) ExecRead(ctx context.Context, fn func(tx driver.ManagedTransaction) (any, error)) (any, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.out, nil
}

func TestReadUsecase_ListCategoryMeta(t *testing.T) {
	uc := NewReadUsecase(&fakeReadExec{})
	meta := uc.ListCategoryMeta()
	if len(meta) < 9 {
		t.Fatalf("categories: %d", len(meta))
	}
	var engage bool
	for _, m := range meta {
		if m.ID == "engage" {
			engage = true
			if len(m.Labels) < 3 {
				t.Fatalf("engage labels: %v", m.Labels)
			}
		}
	}
	if !engage {
		t.Fatal("engage category missing")
	}
}

func TestReadUsecase_GetNodeForAPI(t *testing.T) {
	ctx := context.Background()

	t.Run("not found", func(t *testing.T) {
		uc := NewReadUsecase(&fakeReadExec{out: (*query.Node)(nil)})
		_, err := uc.GetNodeForAPI(ctx, "missing")
		if !errors.Is(err, domain.ErrNodeNotFound) {
			t.Fatalf("err: %v", err)
		}
	})

	t.Run("found", func(t *testing.T) {
		node := &query.Node{ElementID: "el-1", Labels: []string{"IOC"}}
		uc := NewReadUsecase(&fakeReadExec{out: node})
		got, err := uc.GetNodeForAPI(ctx, "el-1")
		if err != nil {
			t.Fatal(err)
		}
		if got.ElementID != "el-1" {
			t.Fatalf("node: %+v", got)
		}
	})
}

func TestReadUsecase_Ping(t *testing.T) {
	uc := NewReadUsecase(&fakeReadExec{})
	if err := uc.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	ucFail := NewReadUsecase(&fakeReadExec{err: errors.New("down")})
	if err := ucFail.Ping(context.Background()); err == nil {
		t.Fatal("expected ping error")
	}
}

func TestReadUsecase_EngageTargetContext(t *testing.T) {
	ectx := &query.EngageTargetContext{
		Host: "example.com",
		Target: &query.Node{
			ElementID: "t1",
			Labels:    []string{"EngageTarget"},
			Props:     map[string]any{"name": "example.com"},
		},
	}
	uc := NewReadUsecase(&fakeReadExec{out: ectx})
	got, err := uc.EngageTargetContext(context.Background(), "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Target == nil || got.Host != "example.com" {
		t.Fatalf("context: %+v", got)
	}
}

func TestReadUsecase_TargetGraphState(t *testing.T) {
	tiNodes := []query.Node{{ElementID: "ioc-1", Labels: []string{"IOC"}}}
	engageCtx := &query.EngageTargetContext{
		Host:   "example.com",
		Target: &query.Node{ElementID: "tgt", Labels: []string{"EngageTarget"}},
	}
	uc := NewReadUsecase(&sequentialExec{
		replies: []any{tiNodes, []query.Node{}, []query.Node{}, engageCtx},
	})

	state, err := uc.TargetGraph(context.Background(), "https://example.com", TargetGraphOpts{
		IncludeEngageContext: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Host != "example.com" {
		t.Fatalf("host: %q", state.Host)
	}
	if len(state.Hits["ti"]) != 1 {
		t.Fatalf("ti hits: %v", state.Hits)
	}
	if !state.EngageFound || state.EngageContext == nil {
		t.Fatalf("engage: found=%v ctx=%v", state.EngageFound, state.EngageContext)
	}
}

type sequentialExec struct {
	replies []any
	err     error
	i       int
}

func (s *sequentialExec) ExecRead(ctx context.Context, fn func(tx driver.ManagedTransaction) (any, error)) (any, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.i >= len(s.replies) {
		return nil, nil
	}
	out := s.replies[s.i]
	s.i++
	return out, nil
}
