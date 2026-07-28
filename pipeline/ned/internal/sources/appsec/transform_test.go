package appsec

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
)

func TestTransformSBOM_osvJSON(t *testing.T) {
	raw := harvest.SBOMOSVRaw{
		OSVID:   "OSV-1",
		CVE:     "CVE-2024-1000",
		RawJSON: `{"id":"OSV-1","affected":[{"package":{"ecosystem":"npm","name":"pkg"}}]}`,
	}
	payload, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	env := &harvest.Envelope{
		SchemaVersion: harvest.CurrentSchemaVersion,
		Source:        harvest.SourceSBOM,
		Kind:          harvest.KindSBOMOSVJSON,
		Payload:       payload,
	}
	out, err := TransformSBOM(context.Background(), env)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Kind != commit.KindSBOMOSVRecord {
		t.Fatalf("got %+v", out)
	}
	var pl commit.SBOMOSVPayload
	if err := json.Unmarshal(out[0].Payload, &pl); err != nil {
		t.Fatal(err)
	}
	if pl.CVE != "CVE-2024-1000" || len(pl.Affected) != 1 {
		t.Fatalf("payload %+v", pl)
	}
}

func TestTransformCoderules_cweRow(t *testing.T) {
	raw := harvest.CoderulesCWERaw{ID: "CWE-79", Name: "XSS", Description: "d", Status: "Stable"}
	payload, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	env := &harvest.Envelope{
		SchemaVersion: harvest.CurrentSchemaVersion,
		Source:        harvest.SourceCoderules,
		Kind:          harvest.KindCoderulesCWERaw,
		Payload:       payload,
	}
	out, err := TransformCoderules(context.Background(), env)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Source != commit.SourceCoderules {
		t.Fatalf("got %+v", out[0])
	}
}

func TestTransformNuclei_templateRaw(t *testing.T) {
	yaml := "id: http-missing\ninfo:\n  name: Missing\n  severity: info\n"
	raw := harvest.NucleiTemplateRaw{Path: "/tmp/http-missing.yaml", RawYAML: yaml}
	payload, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	env := &harvest.Envelope{
		SchemaVersion: harvest.CurrentSchemaVersion,
		Source:        harvest.SourceNuclei,
		Kind:          harvest.KindNucleiTemplateRaw,
		Payload:       payload,
	}
	out, err := TransformNuclei(context.Background(), env)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Kind != commit.KindNucleiTemplate {
		t.Fatalf("kind %s", out[0].Kind)
	}
}
