// Package domain holds cybersecurity playbook skill metadata (Anthropic corpus index).
package domain

// SkillMeta is index metadata for one agentskills.io-style skill.
type SkillMeta struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Domain       string   `json:"domain"`
	Subdomain    string   `json:"subdomain"`
	Description  string   `json:"description"`
	Tags         []string `json:"tags"`
	NISTCSF      []string `json:"nist_csf"`
	Version      string   `json:"version"`
	License      string   `json:"license"`
	AttackIDs    []string `json:"attack_ids"`
	CorpusPath   string   `json:"corpus_path"`
	ExternalPath string   `json:"external_path,omitempty"` // deprecated alias of corpus_path
	BodyChars    int      `json:"body_chars"`
}

// IndexFile is the generated docs/skills-index/cyber-skills.json document.
type IndexFile struct {
	SchemaVersion     int            `json:"schema_version"`
	GeneratedAt       string         `json:"generated_at"`
	Source            string         `json:"source"`
	SourcePath        string         `json:"source_path"`
	SkillCount        int            `json:"skill_count"`
	UniqueAttackIDs   int            `json:"unique_attack_ids"`
	SubdomainCounts   map[string]int `json:"subdomain_counts"`
	Skills            []SkillMeta    `json:"skills"`
}

// SkillDetail includes markdown body for MCP/API responses.
type SkillDetail struct {
	SkillMeta
	Body string `json:"body"`
}
