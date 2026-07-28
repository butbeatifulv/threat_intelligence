package domain

// StepKind classifies a procedure step for agents and catalog linking.
type StepKind string

const (
	StepShell  StepKind = "shell"
	StepManual StepKind = "manual"
	StepTool   StepKind = "tool"
)

// ProcedureStep is one workflow step extracted from SKILL.md.
type ProcedureStep struct {
	Number       int      `json:"number"`
	Title        string   `json:"title"`
	Kind         StepKind `json:"kind"`
	Body         string   `json:"body,omitempty"`
	ToolMentions []string `json:"tool_mentions,omitempty"`
	CatalogTools []string `json:"catalog_tools,omitempty"`
}

// ProcedureSummary is the generated procedures-index entry (no full body).
type ProcedureSummary struct {
	ID           string   `json:"id"`
	Subdomain    string   `json:"subdomain"`
	AttackIDs    []string `json:"attack_ids"`
	NISTCSF      []string `json:"nist_csf"`
	StepCount    int      `json:"step_count"`
	WhenToUse    int      `json:"when_to_use_count"`
	PrereqCount  int      `json:"prereq_count"`
	ToolMentions []string `json:"tool_mentions"`
	CatalogTools []string `json:"catalog_tools"`
	CorpusPath   string   `json:"corpus_path"`
}

// ProcedureSpec is structured procedure content for Knowledge domain.
type ProcedureSpec struct {
	ID           string            `json:"id"`
	Subdomain    string            `json:"subdomain"`
	AttackIDs    []string          `json:"attack_ids"`
	NISTCSF      []string          `json:"nist_csf"`
	WhenToUse    []string          `json:"when_to_use"`
	Prerequisites []string         `json:"prerequisites"`
	Steps        []ProcedureStep   `json:"steps"`
	Scenarios    []string          `json:"scenarios,omitempty"`
	ToolMentions []string          `json:"tool_mentions"`
	CatalogTools []string          `json:"catalog_tools"`
}

// ProceduresIndexFile is docs/skills-index/procedures-index.json.
type ProceduresIndexFile struct {
	SchemaVersion int                `json:"schema_version"`
	GeneratedAt   string             `json:"generated_at"`
	SkillCount    int                `json:"skill_count"`
	Procedures    []ProcedureSummary `json:"procedures"`
}
