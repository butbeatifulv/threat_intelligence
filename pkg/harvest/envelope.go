// Package harvest defines the harvest-layer envelope published to NATS scrape.> subjects.
// Importable from scrape and pipeline only (wire contract).
package harvest

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const CurrentSchemaVersion = 1

const (
	SourceSBOM      = "sbom"
	SourceCoderules = "coderules"
	SourceNuclei    = "nuclei"
	SourceTI        = "ti"
	SourceVuln      = "vuln"
	SourceLola      = "lola"
	SourceDS        = "ds"
	SourceBrowser   = "browser"
)

const (
	KindSBOMOSVJSON           = "scrape_sbom_osv_json"
	KindSBOMGHSAPath          = "scrape_sbom_ghsa_path"
	KindCoderulesCWERaw       = "scrape_coderules_cwe_raw"
	KindCoderulesSemgrepRaw   = "scrape_coderules_semgrep_raw"
	KindCoderulesCodeQLRaw    = "scrape_coderules_codeql_raw"
	KindNucleiTemplateRaw     = "scrape_nuclei_template_raw"
	KindTIKEVRow              = "scrape_ti_kev_row"
	KindTIJSONLLine           = "scrape_ti_jsonl_line"
	KindTIIoCRaw              = "scrape_ti_ioc_raw"
	KindTIReportRaw           = "scrape_ti_report_raw"
	KindTICampaignRaw       = "scrape_ti_campaign_raw"
	KindTIClusterRaw          = "scrape_ti_cluster_raw"
	KindTIActorRaw            = "scrape_ti_actor_raw"
	KindVulnCVEUpsert         = "scrape_vuln_cve_upsert"
	KindVulnMergeExploit      = "scrape_vuln_merge_exploit"
	KindVulnNVDPage           = "scrape_nvd_page"
	KindLolaArtifactRaw       = "scrape_lola_artifact_raw"
	KindLolaLoftsRaw          = "scrape_lola_lofts_raw"
	KindLolaLinkArtifacts     = "scrape_lola_link_artifacts"
	KindLolaAttackTechnique   = "scrape_lola_attack_technique"
	KindLolaAttackTactic      = "scrape_lola_attack_tactic"
	KindLolaMergeTacticTechnique = "scrape_lola_merge_tactic_technique"
	KindLolaMergeSubtechnique = "scrape_lola_merge_subtechnique"
	KindDSSigmaRaw            = "scrape_ds_sigma_raw"
	KindDSYaraRaw             = "scrape_ds_yara_raw"
	KindDSAtomicRaw           = "scrape_ds_atomic_raw"
	KindDSCalderaRaw          = "scrape_ds_caldera_raw"
	KindBrowserInspectRaw     = "scrape_browser_inspect_raw"
)

// Envelope is the on-wire JSON for scrape.> JetStream messages.
type Envelope struct {
	SchemaVersion int             `json:"schema_version"`
	Source        string          `json:"source"`
	Kind          string          `json:"kind"`
	ContentKey    string          `json:"content_key"`
	ScrapedAt     string          `json:"scraped_at"`
	Payload       json.RawMessage `json:"payload"`
}

func (e *Envelope) Validate() error {
	if e.SchemaVersion != CurrentSchemaVersion {
		return fmt.Errorf("harvest: unsupported schema_version %d (want %d)", e.SchemaVersion, CurrentSchemaVersion)
	}
	if strings.TrimSpace(e.Source) == "" {
		return errors.New("harvest: empty source")
	}
	if strings.TrimSpace(e.Kind) == "" {
		return errors.New("harvest: empty kind")
	}
	if strings.TrimSpace(e.ContentKey) == "" {
		return errors.New("harvest: empty content_key")
	}
	if strings.TrimSpace(e.ScrapedAt) == "" {
		return errors.New("harvest: empty scraped_at")
	}
	if len(e.Payload) == 0 || string(e.Payload) == "null" {
		return errors.New("harvest: empty payload")
	}
	return nil
}

// NewEnvelope marshals payload and sets schema metadata.
func NewEnvelope(source, kind, contentKey string, payload any) (*Envelope, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	env := &Envelope{
		SchemaVersion: CurrentSchemaVersion,
		Source:        source,
		Kind:          kind,
		ContentKey:    contentKey,
		ScrapedAt:     time.Now().UTC().Format(time.RFC3339),
		Payload:       data,
	}
	if err := env.Validate(); err != nil {
		return nil, err
	}
	return env, nil
}

// DSContentKey builds a stable ledger/NATS key for ds artifacts.
func DSContentKey(prefix, path string) string {
	return "ds:" + prefix + ":" + strings.TrimSpace(path)
}

type DSSigmaRaw struct {
	Path    string `json:"path"`
	RawYAML string `json:"raw_yaml"`
}

type DSYaraRaw struct {
	Path    string `json:"path"`
	Name    string `json:"name,omitempty"`
	RawBody string `json:"raw_body"`
}

type DSAtomicRaw struct {
	Path    string `json:"path"`
	RawYAML string `json:"raw_yaml"`
}

type DSCalderaRaw struct {
	Path     string `json:"path"`
	FileName string `json:"file_name"`
	RawBody  string `json:"raw_body"`
}

type TIKEVRow struct {
	CVEID         string `json:"cve_id"`
	VendorProject string `json:"vendor_project,omitempty"`
	Product       string `json:"product,omitempty"`
	ShortDesc     string `json:"short_description,omitempty"`
	DateAdded     string `json:"date_added,omitempty"`
}

type TIJSONLLine struct {
	Line json.RawMessage `json:"line"`
}

type SBOMOSVRaw struct {
	CVE     string `json:"cve"`
	OSVID   string `json:"osv_id"`
	RawJSON string `json:"raw_json"`
}

type SBOMGHSARaw struct {
	Path string         `json:"path"`
	Doc  map[string]any `json:"doc"`
}

type CoderulesCWERaw struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type CoderulesSemgrepRaw struct {
	Path    string `json:"path"`
	RawYAML string `json:"raw_yaml"`
}

type CoderulesCodeQLRaw struct {
	Path string `json:"path"`
	Body string `json:"body"`
}

type NucleiTemplateRaw struct {
	Path    string `json:"path"`
	RawYAML string `json:"raw_yaml"`
}

type VulnMergeExploit struct {
	CVE    string `json:"cve"`
	Source string `json:"source"`
	RefID  string `json:"ref_id"`
	URL    string `json:"url,omitempty"`
}

type VulnNVDPage struct {
	StartIndex int    `json:"start_index"`
	RawJSON    string `json:"raw_json"`
}

type LolaArtifactRaw struct {
	Source  string `json:"source"`
	Path    string `json:"path"`
	RawBody string `json:"raw_body"`
}

type LolaLoftsRaw struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	LinkURL  string `json:"link_url"`
	Markdown string `json:"markdown"`
}

type LolaAttackTechnique struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Markdown    string `json:"markdown,omitempty"`
}

type LolaAttackTactic struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Markdown    string `json:"markdown,omitempty"`
}

type LolaMergeTacticTechnique struct {
	TacticID    string `json:"tactic_id"`
	TechniqueID string `json:"technique_id"`
}

type LolaMergeSubtechnique struct {
	ParentTechniqueID string `json:"parent_technique_id"`
	ChildTechniqueID  string `json:"child_technique_id"`
}

// BrowserInspectRaw is the harvest payload for a Playwright inspect crawl.
type BrowserInspectRaw struct {
	URL       string `json:"url"`
	RawJSON   string `json:"raw_json"`
	Timestamp string `json:"timestamp,omitempty"`
}

// BrowserContentKey builds a stable ledger/NATS key for browser inspect results.
func BrowserContentKey(url string) string {
	return "browser:inspect:" + strings.TrimSpace(url)
}
