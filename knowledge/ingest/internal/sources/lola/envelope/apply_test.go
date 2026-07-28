package ingest

import (
	"context"
	"testing"

	neo4jstore "github.com/butbeautifulv/veil/knowledge/ingest/internal/sources/lola/storage"
	"github.com/butbeautifulv/veil/pkg/commit"
)

func TestApplyEnvelope_unknownKind(t *testing.T) {
	env := &commit.Envelope{
		SchemaVersion: commit.CurrentSchemaVersion,
		Source:        commit.SourceLola,
		Kind:          "lola_unknown",
		Payload:       []byte(`{}`),
	}
	err := ApplyEnvelope(context.Background(), (*neo4jstore.Store)(nil), env)
	if err == nil {
		t.Fatal("expected error")
	}
}
