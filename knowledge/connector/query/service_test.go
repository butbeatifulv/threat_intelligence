package query

import (
	"context"
	"errors"
	"testing"

	driver "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type stubExec struct {
	out any
	err error
}

func (s *stubExec) ExecRead(ctx context.Context, fn func(tx driver.ManagedTransaction) (any, error)) (any, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.out, nil
}

func TestService_validationErrors(t *testing.T) {
	svc := NewService(&stubExec{})
	ctx := context.Background()

	tests := []struct {
		name string
		run  func() error
	}{
		{"NodesByKind empty", func() error { _, err := svc.NodesByKind(ctx, "", 10, 0); return err }},
		{"NodesByCategory unknown", func() error { _, err := svc.NodesByCategory(ctx, "nope", "IOC", 10, 0); return err }},
		{"NodesByCategory bad kind", func() error { _, err := svc.NodesByCategory(ctx, "ti", "Vulnerability", 10, 0); return err }},
		{"GetNode empty", func() error { _, err := svc.GetNode(ctx, ""); return err }},
		{"Neighbors empty", func() error { _, err := svc.Neighbors(ctx, "", 1, 10); return err }},
		{"Search empty", func() error { _, err := svc.Search(ctx, "", "", 10); return err }},
		{"SearchInCategory unknown", func() error { _, err := svc.SearchInCategory(ctx, "nope", "x", "", 10); return err }},
		{"SearchInCategory empty query", func() error { _, err := svc.SearchInCategory(ctx, "ti", "", "", 10); return err }},
		{"SearchInCategory bad kind", func() error { _, err := svc.SearchInCategory(ctx, "ti", "cve", "Vulnerability", 10); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestService_NodesByKind_returnsStub(t *testing.T) {
	want := []Node{{ElementID: "n1", Labels: []string{"IOC"}}}
	svc := NewService(&stubExec{out: want})
	nodes, err := svc.NodesByKind(context.Background(), "IOC", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 || nodes[0].ElementID != "n1" {
		t.Fatalf("nodes: %+v", nodes)
	}
}

func TestService_GetNode_nil(t *testing.T) {
	svc := NewService(&stubExec{out: (*Node)(nil)})
	n, err := svc.GetNode(context.Background(), "missing")
	if err != nil {
		t.Fatal(err)
	}
	if n != nil {
		t.Fatalf("want nil, got %+v", n)
	}
}

func TestService_ListKindsInCategory_unknown(t *testing.T) {
	svc := NewService(&stubExec{})
	_, err := svc.ListKindsInCategory(context.Background(), "unknown-cat")
	if err == nil {
		t.Fatal("expected unknown category error")
	}
}

func TestService_execError(t *testing.T) {
	svc := NewService(&stubExec{err: errors.New("neo4j down")})
	_, err := svc.ListKinds(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
