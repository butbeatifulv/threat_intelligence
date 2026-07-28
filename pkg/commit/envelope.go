// Package commit defines the NED → graph envelope published to NATS ingest.> subjects.
// Importable from pipeline and graph only (wire contract).
package commit

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const CurrentSchemaVersion = 1

// Well-known source identifiers (scrapers).
const (
	SourceSBOM      = "sbom"
	SourceCoderules = "coderules"
	SourceNuclei    = "nuclei"
	SourceTI        = "ti"
	SourceVuln      = "vuln"
	SourceLola      = "lola"
	SourceDS        = "ds"
	SourceEngage    = "engage"
)

// Event kinds per source (extend as needed).
const (
	KindSBOMOSVRecord    = "sbom_osv_record"
	KindSBOMGHSADocument = "sbom_ghsa_document"
	KindCoderulesCWERow  = "coderules_cwe_row"
	KindCoderulesSemgrep = "coderules_semgrep_yaml"
	KindCoderulesCodeQL  = "coderules_codeql_ql"
	KindNucleiTemplate   = "nuclei_template_yaml"

	// TI (threat intel feeds / JSONL) — upsert payloads are normalized by NED; graph MERGE uses IdempotencyKey node ids (see ti_node.go).
	KindTIIoC                   = "ti_ioc"
	KindTIKEVVulnerability      = "ti_kev_vulnerability"
	KindTIReport                = "ti_report"
	KindTICampaign              = "ti_campaign"
	KindTICluster               = "ti_cluster"
	KindTIActor                 = "ti_actor"
	KindTILinkCampaignIOC       = "ti_link_campaign_ioc"
	KindTILinkClusterCampaign    = "ti_link_cluster_campaign"
	KindTILinkCampaignActor      = "ti_link_campaign_actor"
	KindTILinkReportMentionsIOC  = "ti_link_report_mentions_ioc"
	KindTIJSONLRecord            = "ti_jsonl_record"

	// vuln — NVD + exploit refs (same semantics as vuln/storage/neo4j).
	KindVulnUpsert        = "vuln_upsert"
	KindVulnMergeExploit  = "vuln_merge_exploit"

	// lola — LOLBAS / GTFOBins / LOFTS / MITRE STIX (repository.LolaRepository).
	KindLolaArtifact              = "lola_artifact"
	KindLolaLofts                 = "lola_lofts"
	KindLolaAttackTechnique       = "lola_attack_technique"
	KindLolaAttackTactic          = "lola_attack_tactic"
	KindLolaMergeTacticTechnique  = "lola_merge_tactic_technique"
	KindLolaMergeSubtechnique     = "lola_merge_subtechnique"
	KindLolaLinkArtifacts         = "lola_link_artifacts"

	// ds — detection content writers (Sigma / YARA / Atomic / Caldera).
	KindDSUpsertSigma    = "ds_upsert_sigma"
	KindDSUpsertYara     = "ds_upsert_yara"
	KindDSUpsertAtomic   = "ds_upsert_atomic"
	KindDSUpsertCaldera  = "ds_upsert_caldera"

	// engage — active security testing tool runs (engage → pipeline ingest).
	KindEngageToolRun = "engage_tool_run"
	KindEngageFinding = "engage_finding"
)

// Envelope is the on-wire JSON for JetStream / HTTP bridges.
type Envelope struct {
	SchemaVersion  int             `json:"schema_version"`
	Source         string          `json:"source"`
	Kind           string          `json:"kind"`
	IdempotencyKey string          `json:"idempotency_key"`
	Payload        json.RawMessage `json:"payload"`
}

// Validate checks required fields for schema v1.
func (e *Envelope) Validate() error {
	if e.SchemaVersion != CurrentSchemaVersion {
		return fmt.Errorf("commit: unsupported schema_version %d (want %d)", e.SchemaVersion, CurrentSchemaVersion)
	}
	if strings.TrimSpace(e.Source) == "" {
		return errors.New("commit: empty source")
	}
	if strings.TrimSpace(e.Kind) == "" {
		return errors.New("commit: empty kind")
	}
	if strings.TrimSpace(e.IdempotencyKey) == "" {
		return errors.New("commit: empty idempotency_key")
	}
	if len(e.Payload) == 0 || string(e.Payload) == "null" {
		return errors.New("commit: empty payload")
	}
	return nil
}

// SBOMOSVPayload is the payload for KindSBOMOSVRecord.
type SBOMOSVPayload struct {
	OSVID    string           `json:"osv_id"`
	CVE      string           `json:"cve"`
	Affected []map[string]any `json:"affected"`
}

// SBOMGHSAPathPayload carries GHSA JSON plus stable path for idempotency.
type SBOMGHSAPathPayload struct {
	Path string         `json:"path"`
	Doc  map[string]any `json:"doc"`
}

// CoderulesCWEPayload is one CWE catalog row.
type CoderulesCWEPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// CoderulesSemgrepPayload is raw YAML body + metadata for storage.
type CoderulesSemgrepPayload struct {
	Path     string   `json:"path"`
	Language string   `json:"language"`
	RuleID   string   `json:"rule_id"`
	Title    string   `json:"title"`
	RawYAML  string   `json:"raw_yaml"`
	CWEs     []string `json:"cwes"`
}

// CoderulesCodeQLPayload is a CodeQL query file snapshot.
type CoderulesCodeQLPayload struct {
	Path string   `json:"path"`
	Name string   `json:"name"`
	Lang string   `json:"lang"`
	Body string   `json:"body"`
	CWEs []string `json:"cwes"`
}

// NucleiTemplatePayload is parsed template fields + raw YAML.
type NucleiTemplatePayload struct {
	Path       string `json:"path"`
	TemplateID string `json:"template_id"`
	Name       string `json:"name"`
	Severity   string `json:"severity"`
	TagsJSON   string `json:"tags_json"`
	CVE        string `json:"cve"`
	CWE        string `json:"cwe"`
	RawYAML    string `json:"raw_yaml"`
}

// TIKEVVulnPayload is one CISA KEV row (same fields as ti feeds runner).
type TIKEVVulnPayload struct {
	CVEID         string `json:"cve_id"`
	VendorProject string `json:"vendor_project"`
	Product       string `json:"product"`
	ShortDesc     string `json:"short_desc"`
	DateAdded     string `json:"date_added"`
}

// TILinkCampaignIOCPayload links Campaign → IOC after both upserts.
type TILinkCampaignIOCPayload struct {
	CampaignID string          `json:"campaign_id"`
	IOC        json.RawMessage `json:"ioc"`
}

// TILinkClusterCampaignPayload links Cluster → Campaign.
type TILinkClusterCampaignPayload struct {
	ClusterID  string `json:"cluster_id"`
	CampaignID string `json:"campaign_id"`
}

// TILinkCampaignActorPayload links Campaign → Actor (actor identified by name).
type TILinkCampaignActorPayload struct {
	CampaignID string `json:"campaign_id"`
	ActorName  string `json:"actor_name"`
}

// TILinkReportMentionsIOCPayload links Report → IOC.
type TILinkReportMentionsIOCPayload struct {
	ReportID string          `json:"report_id"`
	IOC      json.RawMessage `json:"ioc"`
}

// TIJSONLRecordPayload is one raw JSONL line (ti/internal/ingest.Envelope JSON).
type TIJSONLRecordPayload struct {
	Line json.RawMessage `json:"line"`
}

// VulnMergeExploitPayload links an Exploit node to a Vulnerability (CVE must exist).
type VulnMergeExploitPayload struct {
	CVE    string `json:"cve"`
	Source string `json:"source"`
	RefID  string `json:"ref_id"`
	URL    string `json:"url"`
}

// LolaArtifactPayload is one artifact JSON (lola/domain.Artifact) plus feed source key.
type LolaArtifactPayload struct {
	Source string          `json:"source"`
	Body   json.RawMessage `json:"body"`
}

// LolaLoftsPayload is one LOFTS markdown row.
type LolaLoftsPayload struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	LinkURL  string `json:"link_url"`
	Markdown string `json:"markdown"`
}

// LolaAttackTechniquePayload / LolaAttackTacticPayload mirror store upsert args.
type LolaAttackTechniquePayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Markdown    string `json:"markdown"`
}

type LolaAttackTacticPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Markdown    string `json:"markdown"`
}

type LolaMergeTacticTechniquePayload struct {
	TacticID     string `json:"tactic_id"`
	TechniqueID  string `json:"technique_id"`
}

type LolaMergeSubtechniquePayload struct {
	ParentTechniqueID string `json:"parent_technique_id"`
	ChildTechniqueID  string `json:"child_technique_id"`
}

// LolaLinkArtifactsPayload is a marker job (empty object) to run post-STIX linking.
type LolaLinkArtifactsPayload struct{}

// DS payloads mirror neo4j store method parameters.
type DSUpsertSigmaPayload struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Level       string `json:"level"`
	LogProduct  string `json:"log_product"`
	LogService  string `json:"log_service"`
	TagsJSON    string `json:"tags_json"`
	Markdown    string `json:"markdown"`
	Source      string `json:"source"`
}

type DSUpsertYaraPayload struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Author   string `json:"author"`
	TagsJSON string `json:"tags_json"`
	Markdown string `json:"markdown"`
	Source   string `json:"source"`
}

type DSUpsertAtomicPayload struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Tactic     string `json:"tactic"`
	Technique  string `json:"technique"`
	ExecName   string `json:"exec_name"`
	ExecCmd    string `json:"exec_cmd"`
	Markdown   string `json:"markdown"`
	Source     string `json:"source"`
}

type DSUpsertCalderaPayload struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Tactic       string `json:"tactic"`
	TechniqueID  string `json:"technique_id"`
	Markdown     string `json:"markdown"`
	Source       string `json:"source"`
}
func NewEnvelope(source, kind, idempotencyKey string, payload any) (*Envelope, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &Envelope{
		SchemaVersion:  CurrentSchemaVersion,
		Source:         source,
		Kind:           kind,
		IdempotencyKey: idempotencyKey,
		Payload:        raw,
	}, nil
}

// SBOMOSVIdempotencyKey builds a stable key per CVE + package row (ecosystem:name).
func SBOMOSVIdempotencyKey(cve, ecosystem, pkgName string) string {
	cve = strings.TrimSpace(strings.ToUpper(cve))
	return fmt.Sprintf("sbom:osv:%s:%s:%s", cve, strings.ToLower(strings.TrimSpace(ecosystem)), strings.TrimSpace(pkgName))
}

// SBOMGHSAIdempotencyKey uses advisory path in upstream repo.
func SBOMGHSAIdempotencyKey(path string) string {
	return "sbom:ghsa:" + strings.TrimSpace(path)
}

// CoderulesCWEIdempotencyKey is one row from MITRE catalog.
func CoderulesCWEIdempotencyKey(cweID string) string {
	return "coderules:cwe:" + strings.TrimSpace(strings.ToUpper(cweID))
}

// CoderulesSemgrepIdempotencyKey addresses one rule file path in the registry.
func CoderulesSemgrepIdempotencyKey(path string) string {
	return "coderules:semgrep:" + strings.TrimSpace(path)
}

// CoderulesCodeQLIdempotencyKey addresses one .ql path.
func CoderulesCodeQLIdempotencyKey(path string) string {
	return "coderules:codeql:" + strings.TrimSpace(path)
}

// NucleiTemplateIdempotencyKey addresses one template path.
func NucleiTemplateIdempotencyKey(path string) string {
	return "nuclei:template:" + strings.TrimSpace(path)
}

// --- TI idempotency keys (prefix ti:) ---

func TIIoCIdempotencyKey(canonicalIOCID string) string {
	return "ti:ioc:" + strings.TrimSpace(canonicalIOCID)
}

func TIKEVIdempotencyKey(cve string) string {
	return "ti:kev:" + strings.TrimSpace(strings.ToUpper(cve))
}

func TIReportIdempotencyKey(stableReportID string) string {
	return "ti:report:" + strings.TrimSpace(stableReportID)
}

func TICampaignIdempotencyKey(id string) string {
	return "ti:campaign:" + strings.TrimSpace(id)
}

func TIClusterIdempotencyKey(id string) string {
	return "ti:cluster:" + strings.TrimSpace(id)
}

func TIActorIdempotencyKey(actorStableID string) string {
	return "ti:actor:" + strings.TrimSpace(actorStableID)
}

func TILinkCampaignIOCIdempotencyKey(campaignID, iocCanonicalID string) string {
	return "ti:lc:" + strings.TrimSpace(campaignID) + ":" + strings.TrimSpace(iocCanonicalID)
}

func TILinkClusterCampaignIdempotencyKey(clusterID, campaignID string) string {
	return "ti:lkc:" + strings.TrimSpace(clusterID) + ":" + strings.TrimSpace(campaignID)
}

func TILinkCampaignActorIdempotencyKey(campaignID, actorStableID string) string {
	return "ti:lca:" + strings.TrimSpace(campaignID) + ":" + strings.TrimSpace(actorStableID)
}

func TILinkReportMentionsIOCIdempotencyKey(reportID, iocCanonicalID string) string {
	return "ti:lrmi:" + strings.TrimSpace(reportID) + ":" + strings.TrimSpace(iocCanonicalID)
}

func TIJSONLRecordIdempotencyKey(lineSHA256Hex32 string) string {
	return "ti:jsonl:" + strings.TrimSpace(lineSHA256Hex32)
}

// --- vuln ---

func VulnUpsertIdempotencyKey(cve string) string {
	return "vuln:upsert:" + strings.TrimSpace(strings.ToUpper(cve))
}

func VulnMergeExploitIdempotencyKey(cve, source, refID string) string {
	return "vuln:exp:" + strings.TrimSpace(strings.ToUpper(cve)) + ":" + strings.TrimSpace(source) + ":" + strings.TrimSpace(refID)
}

// --- lola ---

func LolaArtifactIdempotencyKey(source, artifactName string) string {
	return "lola:art:" + strings.TrimSpace(strings.ToLower(source)) + ":" + strings.TrimSpace(artifactName)
}

func LolaLoftsIdempotencyKey(linkURL string) string {
	return "lola:lofts:" + strings.TrimSpace(linkURL)
}

func LolaTechniqueIdempotencyKey(id string) string {
	return "lola:tech:" + strings.TrimSpace(id)
}

func LolaTacticIdempotencyKey(id string) string {
	return "lola:tac:" + strings.TrimSpace(id)
}

func LolaMergeTacticTechniqueIdempotencyKey(tac, tech string) string {
	return "lola:mt:" + strings.TrimSpace(tac) + ":" + strings.TrimSpace(tech)
}

func LolaMergeSubtechniqueIdempotencyKey(parent, child string) string {
	return "lola:ms:" + strings.TrimSpace(parent) + ":" + strings.TrimSpace(child)
}

func LolaLinkArtifactsIdempotencyKey() string { return "lola:link:artifacts:v1" }

// --- ds ---

func DSSigmaIdempotencyKey(id string) string   { return "ds:sigma:" + strings.TrimSpace(id) }
func DSYaraIdempotencyKey(id string) string     { return "ds:yara:" + strings.TrimSpace(id) }
func DSAtomicIdempotencyKey(id string) string   { return "ds:atomic:" + strings.TrimSpace(id) }
func DSCalderaIdempotencyKey(id string) string  { return "ds:caldera:" + strings.TrimSpace(id) }

// --- engage ---

// EngageToolRunPayload is published when engage records a tool execution (cross-layer NATS).
type EngageToolRunPayload struct {
	Tool    string `json:"tool"`
	Target  string `json:"target"`
	Subject string `json:"subject"`
	Success bool   `json:"success"`
	At      string `json:"at"` // RFC3339
}

func EngageToolRunIdempotencyKey(tool, target, at string) string {
	return "engage:run:" + strings.TrimSpace(tool) + ":" + strings.TrimSpace(target) + ":" + strings.TrimSpace(at)
}

// EngageFindingPayload is a vulnerability finding from engage smart-scan / assessment.
type EngageFindingPayload struct {
	Tool        string `json:"tool"`
	Target      string `json:"target"`
	Title       string `json:"title"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

func EngageFindingIdempotencyKey(tool, target, title string) string {
	return "engage:finding:" + strings.TrimSpace(tool) + ":" + strings.TrimSpace(target) + ":" + strings.TrimSpace(title)
}
