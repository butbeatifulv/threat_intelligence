package ingest

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"

	"github.com/butbeautifulv/veil/pkg/vuln/domain"
)

type fakeVulnRepo struct {
	upserted []*domain.Vulnerability
	merged   []struct {
		cve string
		ref domain.ExploitRef
	}
}

func (f *fakeVulnRepo) Save(ctx context.Context, v *domain.Vulnerability) error {
	return f.Upsert(ctx, v)
}

func (f *fakeVulnRepo) Upsert(ctx context.Context, v *domain.Vulnerability) error {
	f.upserted = append(f.upserted, v)
	return nil
}

func (f *fakeVulnRepo) MergeExploitForCVE(ctx context.Context, cve string, ref domain.ExploitRef) error {
	f.merged = append(f.merged, struct {
		cve string
		ref domain.ExploitRef
	}{cve, ref})
	return nil
}

func TestApplyEnvelope_vulnUpsert_mapsToRepo(t *testing.T) {
	v := domain.Vulnerability{
		CVE:     "CVE-2024-1234",
		ID:      "CVE-2024-1234",
		Summary: "test summary",
		CWE:     []string{"CWE-79"},
	}
	payload, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	env := &commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceVuln,
		Kind:           commit.KindVulnUpsert,
		IdempotencyKey: commit.VulnUpsertIdempotencyKey(v.CVE),
		Payload:        payload,
	}
	fake := &fakeVulnRepo{}
	if err := ApplyEnvelope(context.Background(), fake, env); err != nil {
		t.Fatal(err)
	}
	if len(fake.upserted) != 1 {
		t.Fatalf("upserted %d", len(fake.upserted))
	}
	got := fake.upserted[0]
	if got.CVE != v.CVE || got.Summary != v.Summary || len(got.CWE) != 1 || got.CWE[0] != "CWE-79" {
		t.Fatalf("got %+v", got)
	}
}

func TestApplyEnvelope_mergeExploit_mapsPayload(t *testing.T) {
	payload, _ := json.Marshal(commit.VulnMergeExploitPayload{
		CVE: "CVE-2024-9999", Source: "exploitdb", RefID: "EDB-1", URL: "https://example.com",
	})
	env := &commit.Envelope{
		SchemaVersion:  commit.CurrentSchemaVersion,
		Source:         commit.SourceVuln,
		Kind:           commit.KindVulnMergeExploit,
		IdempotencyKey: commit.VulnMergeExploitIdempotencyKey("CVE-2024-9999", "exploitdb", "EDB-1"),
		Payload:        payload,
	}
	fake := &fakeVulnRepo{}
	if err := ApplyEnvelope(context.Background(), fake, env); err != nil {
		t.Fatal(err)
	}
	if len(fake.merged) != 1 {
		t.Fatalf("merged %d", len(fake.merged))
	}
	if fake.merged[0].cve != "CVE-2024-9999" || fake.merged[0].ref.Source != "exploitdb" || fake.merged[0].ref.RefID != "EDB-1" {
		t.Fatalf("merged %+v", fake.merged[0])
	}
}

func TestApplyEnvelope_unknownKind(t *testing.T) {
	err := ApplyEnvelope(context.Background(), &fakeVulnRepo{}, &commit.Envelope{Kind: "unknown"})
	if err == nil {
		t.Fatal("expected error")
	}
}
