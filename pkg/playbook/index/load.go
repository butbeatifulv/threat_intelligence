package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/butbeautifulv/veil/pkg/playbook/domain"
)

const (
	DefaultIndexRel     = "docs/skills-index/cyber-skills.json"
	DefaultSkillsRel    = "corpus/anthropic-cybersecurity-skills/skills"
	DefaultMappingsRel  = "pkg/playbook/corpus/mappings"
	MaxBodyBytes        = 64 << 10
	EnvIndexPath        = "VEIL_CYBER_SKILLS_INDEX"
	EnvRepoRoot         = "VEIL_REPO_ROOT"
)

// Catalog loads and searches the generated skills index.
type Catalog struct {
	repoRoot string
	file     domain.IndexFile
	byID     map[string]domain.SkillMeta
}

var (
	defaultCat   *Catalog
	defaultOnce  sync.Once
	defaultErr   error
	osGetwd = os.Getwd
)

// SetRepoGetwd overrides Getwd for RepoRoot discovery; returns restore.
func SetRepoGetwd(fn func() (string, error)) func() {
	old := osGetwd
	osGetwd = fn
	return func() { osGetwd = old }
}

// Default opens the catalog once (lazy). Safe for concurrent read after load.
func Default() (*Catalog, error) {
	defaultOnce.Do(func() {
		defaultCat, defaultErr = Open("")
	})
	return defaultCat, defaultErr
}

// Open loads index from path (empty = env or default under repo root).
func Open(indexPath string) (*Catalog, error) {
	root, err := RepoRoot()
	if err != nil {
		return nil, err
	}
	if indexPath == "" {
		indexPath = os.Getenv(EnvIndexPath)
	}
	if indexPath == "" {
		indexPath = filepath.Join(root, DefaultIndexRel)
	} else if !filepath.IsAbs(indexPath) {
		indexPath = filepath.Join(root, indexPath)
	}
	raw, err := os.ReadFile(indexPath) // #nosec G304 -- indexPath is a deploy-time env var/config value with a committed-file default, not network input
	if err != nil {
		return nil, fmt.Errorf("playbook index: read %s: %w", indexPath, err)
	}
	var file domain.IndexFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("playbook index: decode: %w", err)
	}
	byID := make(map[string]domain.SkillMeta, len(file.Skills))
	for _, s := range file.Skills {
		byID[s.ID] = s
	}
	return &Catalog{repoRoot: root, file: file, byID: byID}, nil
}

// RepoRoot finds Veil repository root (versions.env + go.mod).
func RepoRoot() (string, error) {
	if r := strings.TrimSpace(os.Getenv(EnvRepoRoot)); r != "" {
		return r, nil
	}
	wd, err := osGetwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for i := 0; i < 12; i++ {
		if _, err := os.Stat(filepath.Join(dir, "versions.env")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return wd, nil
}

func (c *Catalog) Meta() domain.IndexFile { return c.file }

func (c *Catalog) Get(id string) (domain.SkillMeta, bool) {
	s, ok := c.byID[id]
	return s, ok
}

// ReadBody loads SKILL.md body for a skill id.
func (c *Catalog) ReadBody(id string) (domain.SkillDetail, error) {
	meta, ok := c.byID[id]
	if !ok {
		return domain.SkillDetail{}, fmt.Errorf("playbook: unknown skill %q", id)
	}
	path := filepath.Join(c.repoRoot, filepath.FromSlash(skillMarkdownRel(meta)))
	raw, err := os.ReadFile(path) // #nosec G304 -- meta comes from a map lookup constrained to ids loaded from the committed skills index, not attacker-controlled
	if err != nil {
		return domain.SkillDetail{}, fmt.Errorf("playbook: read %s: %w", path, err)
	}
	if len(raw) > MaxBodyBytes {
		raw = raw[:MaxBodyBytes]
	}
	body := string(raw)
	if i := strings.Index(body, "---\n"); i >= 0 {
		if j := strings.Index(body[i+4:], "\n---\n"); j >= 0 {
			body = body[i+4+j+5:]
		}
	}
	return domain.SkillDetail{SkillMeta: meta, Body: body}, nil
}

// Search scores skills by token overlap on name, description, tags, subdomain.
func (c *Catalog) Search(query, subdomain string, limit int) []domain.SkillMeta {
	if limit <= 0 {
		limit = 20
	}
	q := strings.ToLower(strings.TrimSpace(query))
	sub := strings.ToLower(strings.TrimSpace(subdomain))
	type scored struct {
		s     domain.SkillMeta
		score int
	}
	var hits []scored
	tokens := tokenize(q)
	for _, s := range c.file.Skills {
		if sub != "" && strings.ToLower(s.Subdomain) != sub {
			continue
		}
		hay := strings.ToLower(s.Name + " " + s.Description + " " + s.Subdomain + " " + strings.Join(s.Tags, " "))
		sc := 0
		if q != "" {
			if strings.Contains(hay, q) {
				sc += 10
			}
			for _, t := range tokens {
				if len(t) >= 3 && strings.Contains(hay, t) {
					sc++
				}
			}
		} else {
			sc = 1
		}
		if sc > 0 {
			hits = append(hits, scored{s: s, score: sc})
		}
	}
	for i := 0; i < len(hits); i++ {
		for j := i + 1; j < len(hits); j++ {
			if hits[j].score > hits[i].score {
				hits[i], hits[j] = hits[j], hits[i]
			}
		}
	}
	out := make([]domain.SkillMeta, 0, limit)
	for i := 0; i < len(hits) && i < limit; i++ {
		out = append(out, hits[i].s)
	}
	return out
}

// ByTechnique returns skills whose attack_ids contain techniqueID (case-insensitive).
func (c *Catalog) ByTechnique(techniqueID string) []domain.SkillMeta {
	tid := strings.ToUpper(strings.TrimSpace(techniqueID))
	var out []domain.SkillMeta
	for _, s := range c.file.Skills {
		for _, a := range s.AttackIDs {
			if strings.EqualFold(a, tid) {
				out = append(out, s)
				break
			}
		}
	}
	return out
}

func skillMarkdownRel(meta domain.SkillMeta) string {
	if p := strings.TrimSpace(meta.CorpusPath); p != "" {
		return p
	}
	return meta.ExternalPath
}

// MappingsDir returns committed framework mappings (MITRE layer, NIST, OWASP).
func MappingsDir() (string, error) {
	root, err := RepoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, DefaultMappingsRel), nil
}

func tokenize(q string) []string {
	var out []string
	for _, p := range strings.FieldsFunc(q, func(r rune) bool {
		return r <= ' ' || r == '-' || r == '_'
	}) {
		if len(p) >= 2 {
			out = append(out, p)
		}
	}
	return out
}
