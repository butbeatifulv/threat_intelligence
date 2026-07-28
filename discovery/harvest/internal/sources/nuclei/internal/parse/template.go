package parse

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParsedTemplate is extracted metadata from a Nuclei YAML document.
type ParsedTemplate struct {
	TemplateID, Name, Severity, TagsJSON, CVE, CWE string
}

// ParseYAML extracts template fields from raw Nuclei YAML bytes.
func ParseYAML(raw []byte) (ParsedTemplate, error) {
	var root map[string]any
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return ParsedTemplate{}, err
	}
	tplID, _ := root["id"].(string)
	tagsJSON := "[]"
	var name, severity, cveID, cweID string
	info, _ := root["info"].(map[string]any)
	if info != nil {
		name, _ = info["name"].(string)
		severity, _ = info["severity"].(string)
		if tg, ok := info["tags"]; ok {
			tags := normalizeTags(tg)
			b, _ := json.Marshal(tags)
			tagsJSON = string(b)
			if tagsJSON == "null" {
				tagsJSON = "[]"
			}
		}
		class, _ := info["classification"].(map[string]any)
		if class != nil {
			if s, ok := class["cve-id"].(string); ok {
				cveID = strings.TrimSpace(strings.ToUpper(s))
			}
			if s, ok := class["cwe-id"].(string); ok {
				cweID = strings.TrimSpace(s)
			}
		}
	}
	if cveID == "" && strings.HasPrefix(strings.ToUpper(tplID), "CVE-") {
		cveID = strings.ToUpper(tplID)
	}
	return ParsedTemplate{
		TemplateID: tplID,
		Name:       name,
		Severity:   severity,
		TagsJSON:   tagsJSON,
		CVE:        cveID,
		CWE:        cweID,
	}, nil
}

func normalizeTags(v any) []string {
	switch t := v.(type) {
	case string:
		var out []string
		for _, p := range strings.Split(t, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	case []any:
		var out []string
		for _, x := range t {
			if s, ok := x.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, strings.TrimSpace(s))
			}
		}
		return out
	default:
		return nil
	}
}
