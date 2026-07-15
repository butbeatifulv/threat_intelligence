package framework

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	pbindex "github.com/butbeautifulv/veil/pkg/playbook/index"
)

// SubdomainEntry describes one Anthropic skill subdomain in the Veil ontology.
type SubdomainEntry struct {
	ID         string   `json:"id"`
	SkillCount int      `json:"skill_count"`
	Priority   string   `json:"priority"`
	VeilCats   []string `json:"veil_categories"`
}

var subdomainVeilCats = map[string][]string{
	"digital-forensics":         {"playbook", "mitre"},
	"incident-response":         {"playbook", "mitre"},
	"threat-hunting":            {"playbook", "detection", "mitre"},
	"malware-analysis":          {"playbook", "mitre"},
	"penetration-testing":       {"playbook", "engage"},
	"web-application-security":  {"playbook", "engage", "mitre"},
	"vulnerability-management": {"playbook", "vuln"},
	"threat-intelligence":       {"playbook", "ti"},
	"cloud-security":            {"playbook", "engage"},
	"detection-engineering":     {"playbook", "detection"},
	"soc-operations":          {"playbook", "engage"},
}

var subdomainPriority = map[string]string{
	"digital-forensics": "P1", "incident-response": "P1", "threat-hunting": "P1",
	"malware-analysis": "P2", "penetration-testing": "P2", "web-application-security": "P2",
	"vulnerability-management": "P2", "threat-intelligence": "P2",
}

var frameworkRepoRoot = pbindex.RepoRoot

// LoadSubdomains builds registry from cyber-skills.json subdomain_counts.
func LoadSubdomains() ([]SubdomainEntry, error) {
	root, err := frameworkRepoRoot()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(root, pbindex.DefaultIndexRel)
	raw, err := os.ReadFile(path) // #nosec G304 -- root is the discovered repo root, DefaultIndexRel is a package constant; no variable path segment
	if err != nil {
		return nil, err
	}
	var doc struct {
		SubdomainCounts map[string]int `json:"subdomain_counts"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	var out []SubdomainEntry
	for id, n := range doc.SubdomainCounts {
		pri := subdomainPriority[id]
		if pri == "" {
			pri = "P3"
		}
		cats := subdomainVeilCats[id]
		if len(cats) == 0 {
			cats = []string{"playbook"}
		}
		out = append(out, SubdomainEntry{
			ID: id, SkillCount: n, Priority: pri, VeilCats: cats,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].SkillCount != out[j].SkillCount {
			return out[i].SkillCount > out[j].SkillCount
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

// SkillsForTechnique returns skill ids from index with attack_ids containing techniqueID.
func SkillsForTechnique(techniqueID string) ([]string, error) {
	techniqueID = strings.ToUpper(strings.TrimSpace(techniqueID))
	root, err := frameworkRepoRoot()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(root, pbindex.DefaultIndexRel)
	raw, err := os.ReadFile(path) // #nosec G304 -- root is the discovered repo root, DefaultIndexRel is a package constant; no variable path segment
	if err != nil {
		return nil, err
	}
	var doc struct {
		Skills []struct {
			ID        string   `json:"id"`
			AttackIDs []string `json:"attack_ids"`
		} `json:"skills"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	var out []string
	for _, s := range doc.Skills {
		for _, tid := range s.AttackIDs {
			if strings.EqualFold(tid, techniqueID) {
				out = append(out, s.ID)
				break
			}
		}
	}
	sort.Strings(out)
	return out, nil
}
