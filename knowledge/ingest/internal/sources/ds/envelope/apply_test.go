package ingest

import (
	"context"
	"testing"

	neo4jstore "github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/ds/storage"
	"github.com/butbeautifulv/veil/pkg/commit"
)

func TestApplyEnvelope_unknownKind(t *testing.T) {
	env := &commit.Envelope{
		SchemaVersion: commit.CurrentSchemaVersion,
		Source:        commit.SourceDS,
		Kind:          "ds_unknown",
		Payload:       []byte(`{}`),
	}
	var st *neo4jstore.Store
	err := ApplyEnvelope(context.Background(), st, env)
	if err == nil {
		t.Fatal("expected error")
	}
}
