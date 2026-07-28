// Package parse extracts vulnerabilities from NVD CVE API 2.0 JSON pages.
package parse

import "encoding/json"

// ParsePage extracts vulnerabilities from an NVD CVE API 2.0 JSON page.
func ParsePage(data []byte) ([]Vulnerability, int, error) {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, 0, err
	}

	items, _ := raw["vulnerabilities"].([]any)
	total := 0
	if tr, ok := raw["totalResults"].(float64); ok {
		total = int(tr)
	}

	var out []Vulnerability
	for _, it := range items {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		cveBlock, _ := m["cve"].(map[string]any)
		id, _ := cveBlock["id"].(string)
		if id == "" {
			continue
		}

		desc := englishDescription(cveBlock)
		cwes := weaknessesToCWE(cveBlock)
		cpes := configurationsToCPE(cveBlock)
		var cvss *CVSS
		if metrics, ok := cveBlock["metrics"].(map[string]any); ok {
			cvss = pickCVSS(metrics, "cvssMetricV31")
			if cvss == nil {
				cvss = pickCVSS(metrics, "cvssMetricV30")
			}
		}

		out = append(out, Vulnerability{
			ID:      id,
			CVE:     id,
			Summary: desc,
			CWE:     cwes,
			CPEs:    cpes,
			CVSS:    cvss,
		})
	}
	return out, total, nil
}

func englishDescription(cveBlock map[string]any) string {
	descs, ok := cveBlock["descriptions"].([]any)
	if !ok {
		return ""
	}
	for _, d := range descs {
		dm, ok := d.(map[string]any)
		if !ok {
			continue
		}
		if lang, _ := dm["lang"].(string); lang == "en" {
			if v, _ := dm["value"].(string); v != "" {
				return v
			}
		}
	}
	if len(descs) > 0 {
		if dm, ok := descs[0].(map[string]any); ok {
			v, _ := dm["value"].(string)
			return v
		}
	}
	return ""
}

func weaknessesToCWE(cveBlock map[string]any) []string {
	var cwes []string
	weaknesses, ok := cveBlock["weaknesses"].([]any)
	if !ok {
		return cwes
	}
	for _, w := range weaknesses {
		wm, ok := w.(map[string]any)
		if !ok {
			continue
		}
		descs, ok := wm["description"].([]any)
		if !ok {
			continue
		}
		for _, dd := range descs {
			dm, ok := dd.(map[string]any)
			if !ok {
				continue
			}
			if v, _ := dm["value"].(string); v != "" {
				cwes = append(cwes, v)
			}
		}
	}
	return cwes
}

func configurationsToCPE(cveBlock map[string]any) []CPE {
	var cpes []CPE
	confs, ok := cveBlock["configurations"].([]any)
	if !ok {
		return cpes
	}
	for _, c := range confs {
		cm, ok := c.(map[string]any)
		if !ok {
			continue
		}
		nodes, ok := cm["nodes"].([]any)
		if !ok {
			continue
		}
		for _, n := range nodes {
			nm, ok := n.(map[string]any)
			if !ok {
				continue
			}
			matches, ok := nm["cpeMatch"].([]any)
			if !ok {
				continue
			}
			for _, mm := range matches {
				mmm, ok := mm.(map[string]any)
				if !ok {
					continue
				}
				if uri, _ := mmm["criteria"].(string); uri != "" {
					cpes = append(cpes, CPE{URI: uri})
				} else if uri2, _ := mmm["cpe23Uri"].(string); uri2 != "" {
					cpes = append(cpes, CPE{URI: uri2})
				}
			}
		}
	}
	return cpes
}

func pickCVSS(metrics map[string]any, key string) *CVSS {
	ms, ok := metrics[key].([]any)
	if !ok || len(ms) == 0 {
		return nil
	}
	m0, ok := ms[0].(map[string]any)
	if !ok {
		return nil
	}
	cv, _ := m0["cvssData"].(map[string]any)
	if cv == nil {
		return nil
	}
	ver, _ := cv["version"].(string)
	vec, _ := cv["vectorString"].(string)
	base, _ := cv["baseScore"].(float64)
	if ver == "" && vec == "" && base == 0 {
		return nil
	}
	return &CVSS{Version: ver, Base: base, Vector: vec}
}
