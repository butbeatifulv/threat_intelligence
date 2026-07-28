package ds

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/butbeautifulv/veil/pkg/commit"
	"github.com/butbeautifulv/veil/pkg/harvest"
	"gopkg.in/yaml.v3"
)

// Transform maps harvest ds events to commit envelopes.
func Transform(ctx context.Context, env *harvest.Envelope) ([]*commit.Envelope, error) {
	_ = ctx
	switch env.Kind {
	case harvest.KindDSSigmaRaw:
		var raw harvest.DSSigmaRaw
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		var root map[string]any
		if err := yaml.Unmarshal([]byte(raw.RawYAML), &root); err != nil {
			return nil, err
		}
		id, _ := root["id"].(string)
		title, _ := root["title"].(string)
		level, _ := root["level"].(string)
		var logProduct, logService string
		if ls, ok := root["logsource"].(map[string]any); ok {
			logProduct, _ = ls["product"].(string)
			logService, _ = ls["service"].(string)
		}
		tags := tagsToJSON(root["tags"])
		if id == "" {
			id = stableID("sigma", raw.Path)
		}
		md := fmt.Sprintf("# %s\n\n**id:** `%s`  \n**level:** %s  \n\n```yaml\n%s\n```\n", title, id, level, raw.RawYAML)
		pl := commit.DSUpsertSigmaPayload{
			ID: id, Title: title, Level: level, LogProduct: logProduct, LogService: logService,
			TagsJSON: tags, Markdown: md, Source: "sigmahq",
		}
		out, err := commit.NewEnvelope(commit.SourceDS, commit.KindDSUpsertSigma, commit.DSSigmaIdempotencyKey(id), pl)
		return []*commit.Envelope{out}, err

	case harvest.KindDSYaraRaw:
		var raw harvest.DSYaraRaw
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		name := raw.Name
		if name == "" {
			name = parseYaraRuleName(raw.RawBody, raw.Path)
		}
		id := stableID("yara", raw.Path)
		md := fmt.Sprintf("# %s\n\n```yara\n%s\n```\n", name, raw.RawBody)
		pl := commit.DSUpsertYaraPayload{ID: id, Name: name, Author: "", TagsJSON: "[]", Markdown: md, Source: "neo23x0-signature-base"}
		out, err := commit.NewEnvelope(commit.SourceDS, commit.KindDSUpsertYara, commit.DSYaraIdempotencyKey(id), pl)
		return []*commit.Envelope{out}, err

	case harvest.KindDSAtomicRaw:
		var raw harvest.DSAtomicRaw
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		return atomicFromYAML(raw.Path, raw.RawYAML)

	case harvest.KindDSCalderaRaw:
		var raw harvest.DSCalderaRaw
		if err := json.Unmarshal(env.Payload, &raw); err != nil {
			return nil, err
		}
		return calderaFromBody(raw.Path, raw.FileName, raw.RawBody)

	default:
		return nil, fmt.Errorf("pipeline ds: unknown kind %q", env.Kind)
	}
}

func stableID(prefix, key string) string {
	h := fmt.Sprintf("%s:%s", prefix, key)
	var x uint64 = 14695981039346656037
	for _, b := range []byte(h) {
		x ^= uint64(b)
		x *= 1099511628211
	}
	return fmt.Sprintf("%s:%016x", prefix, x)
}

func tagsToJSON(v any) string {
	arr, ok := v.([]any)
	if !ok {
		return "[]"
	}
	ss := make([]string, 0, len(arr))
	for _, x := range arr {
		if s, ok := x.(string); ok {
			ss = append(ss, s)
		}
	}
	b, _ := json.Marshal(ss)
	return string(b)
}

func parseYaraRuleName(body, fallback string) string {
	lines := strings.Split(body, "\n")
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if strings.HasPrefix(ln, "rule ") {
			p := strings.TrimPrefix(ln, "rule ")
			if idx := strings.IndexAny(p, " \t{"); idx > 0 {
				return strings.TrimSpace(p[:idx])
			}
			return strings.TrimSpace(p)
		}
	}
	return strings.TrimSuffix(fallback, ".yar")
}

func atomicFromYAML(path, rawYAML string) ([]*commit.Envelope, error) {
	var root map[string]any
	if err := yaml.Unmarshal([]byte(rawYAML), &root); err != nil {
		return nil, err
	}
	attackID, _ := root["attack_technique"].(string)
	atomicTests, _ := root["atomic_tests"].([]any)
	if len(atomicTests) == 0 {
		return nil, fmt.Errorf("no atomic_tests in %s", path)
	}
	var out []*commit.Envelope
	for i, t := range atomicTests {
		tm, ok := t.(map[string]any)
		if !ok {
			continue
		}
		tid, _ := tm["auto_generated_guid"].(string)
		if tid == "" {
			tid = fmt.Sprintf("%s-%d", attackID, i)
		}
		tname, _ := tm["name"].(string)
		tactic := ""
		if ta, ok := tm["tactics"].([]any); ok && len(ta) > 0 {
			if s, ok := ta[0].(string); ok {
				tactic = s
			}
		}
		execName, execCmd := "", ""
		if ex, ok := tm["executor"].(map[string]any); ok {
			execName, _ = ex["name"].(string)
			execCmd, _ = ex["command"].(string)
		}
		md := fmt.Sprintf("# %s\n\n**Technique:** %s  \n**Test:** %s  \n\n```yaml\n%s\n```\n", tname, attackID, tid, rawYAML)
		pl := commit.DSUpsertAtomicPayload{
			ID: tid, Name: tname, Tactic: tactic, Technique: attackID, ExecName: execName, ExecCmd: execCmd, Markdown: md, Source: "atomic-red-team",
		}
		env, err := commit.NewEnvelope(commit.SourceDS, commit.KindDSUpsertAtomic, commit.DSAtomicIdempotencyKey(tid), pl)
		if err != nil {
			return nil, err
		}
		out = append(out, env)
	}
	return out, nil
}

func calderaFromBody(path, _ string, body string) ([]*commit.Envelope, error) {
	var seq []map[string]any
	if err := yaml.Unmarshal([]byte(body), &seq); err == nil && len(seq) > 0 {
		var out []*commit.Envelope
		for _, root := range seq {
			if env := calderaRootToEnvelope(root, path); env != nil {
				out = append(out, env)
			}
		}
		if len(out) > 0 {
			return out, nil
		}
	}
	var root map[string]any
	if err := yaml.Unmarshal([]byte(body), &root); err != nil {
		return nil, err
	}
	if env := calderaRootToEnvelope(root, path); env != nil {
		return []*commit.Envelope{env}, nil
	}
	return nil, nil
}

func calderaRootToEnvelope(root map[string]any, path string) *commit.Envelope {
	id, _ := root["id"].(string)
	name, _ := root["name"].(string)
	desc, _ := root["description"].(string)
	tactic, _ := root["tactic"].(string)
	techID := ""
	if tm, ok := root["technique"].(map[string]any); ok {
		techID, _ = tm["attack_id"].(string)
	}
	if id == "" {
		return nil
	}
	md := fmt.Sprintf("# %s\n\n**Tactic:** %s  \n**Technique:** %s  \n\n%s\n", name, tactic, techID, desc)
	pl := commit.DSUpsertCalderaPayload{ID: id, Name: name, Tactic: tactic, TechniqueID: techID, Markdown: md, Source: "mitre-stockpile"}
	env, err := commit.NewEnvelope(commit.SourceDS, commit.KindDSUpsertCaldera, commit.DSCalderaIdempotencyKey(id), pl)
	if err != nil {
		return nil
	}
	_ = path
	return env
}
