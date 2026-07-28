package parse

import (
	"regexp"
	"strings"
)

func SemgrepMeta(root map[string]any, fileName string) (id, title string) {
	if v, ok := root["id"].(string); ok && v != "" {
		id = v
	}
	if rules, ok := root["rules"].([]any); ok && len(rules) > 0 {
		if rm, ok := rules[0].(map[string]any); ok {
			if id == "" {
				if s, ok := rm["id"].(string); ok {
					id = s
				}
			}
			if msg, ok := rm["message"].(string); ok && strings.TrimSpace(msg) != "" {
				title = strings.TrimSpace(msg)
			}
		}
	}
	if title == "" {
		title = firstNonEmpty(id, fileName)
	}
	if id == "" {
		id = title
	}
	return id, title
}

var cweTokenRe = regexp.MustCompile(`(?i)CWE-\d+`)

func SemgrepCWES(root map[string]any) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(s string) {
		for _, m := range cweTokenRe.FindAllString(s, -1) {
			u := strings.ToUpper(m)
			if _, ok := seen[u]; ok {
				continue
			}
			seen[u] = struct{}{}
			out = append(out, u)
		}
	}
	walkMeta := func(meta map[string]any) {
		if meta == nil {
			return
		}
		if v, ok := meta["cwe"]; ok {
			switch t := v.(type) {
			case string:
				add(t)
			case []any:
				for _, x := range t {
					if s, ok := x.(string); ok {
						add(s)
					}
				}
			}
		}
	}
	if rules, ok := root["rules"].([]any); ok {
		for _, r := range rules {
			rm, ok := r.(map[string]any)
			if !ok {
				continue
			}
			if md, ok := rm["metadata"].(map[string]any); ok {
				walkMeta(md)
			}
		}
	}
	return out
}

func CodeQLCWES(body string) []string {
	lines := strings.Split(body, "\n")
	n := len(lines)
	if n > 100 {
		n = 100
	}
	chunk := strings.Join(lines[:n], "\n")
	seen := map[string]struct{}{}
	var out []string
	for _, m := range cweTokenRe.FindAllString(chunk, -1) {
		u := strings.ToUpper(m)
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}
