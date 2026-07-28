package procedure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	pbindex "github.com/butbeautifulv/veil/pkg/playbook/index"
	"github.com/butbeautifulv/veil/pkg/playbook/cataloglink"
	"github.com/butbeautifulv/veil/pkg/playbook/domain"
)

const DefaultProceduresRel = "docs/skills-index/procedures-index.json"

// Catalog loads procedure summaries and can parse full SKILL.md on demand.
type Catalog struct {
	repoRoot string
	file     domain.ProceduresIndexFile
	byID     map[string]domain.ProcedureSummary
}

var (
	defaultCat  *Catalog
	defaultOnce sync.Once
	defaultErr  error
)

func Default() (*Catalog, error) {
	defaultOnce.Do(func() {
		defaultCat, defaultErr = Open("")
	})
	return defaultCat, defaultErr
}

func Open(path string) (*Catalog, error) {
	root, err := pbindex.RepoRoot()
	if err != nil {
		return nil, err
	}
	if path == "" {
		path = filepath.Join(root, DefaultProceduresRel)
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	raw, err := os.ReadFile(path) // #nosec G304 -- path defaults to a committed index file under repoRoot; caller-supplied override is a config value, not network input
	if err != nil {
		return nil, fmt.Errorf("procedure index: read %s: %w", path, err)
	}
	var file domain.ProceduresIndexFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("procedure index: decode: %w", err)
	}
	byID := make(map[string]domain.ProcedureSummary, len(file.Procedures))
	for _, p := range file.Procedures {
		byID[p.ID] = p
	}
	return &Catalog{repoRoot: root, file: file, byID: byID}, nil
}

func (c *Catalog) Meta() domain.ProceduresIndexFile { return c.file }

func (c *Catalog) GetSummary(id string) (domain.ProcedureSummary, bool) {
	p, ok := c.byID[id]
	return p, ok
}

// GetSpec loads structured procedure (parse SKILL.md + catalog resolve).
func (c *Catalog) GetSpec(id string) (domain.ProcedureSpec, error) {
	sum, ok := c.byID[id]
	if !ok {
		return domain.ProcedureSpec{}, fmt.Errorf("procedure: unknown skill %q", id)
	}
	path := filepath.Join(c.repoRoot, filepath.FromSlash(sum.CorpusPath))
	raw, err := os.ReadFile(path) // #nosec G304 -- sum.CorpusPath is a map lookup constrained to keys loaded from the committed procedures-index.json, not attacker-controlled
	if err != nil {
		return domain.ProcedureSpec{}, fmt.Errorf("procedure: read %s: %w", path, err)
	}
	spec := ParseSkillMarkdown(id, sum.Subdomain, sum.AttackIDs, sum.NISTCSF, string(raw))
	spec.CatalogTools = cataloglink.ResolveMentions(spec.ToolMentions)
	for i := range spec.Steps {
		spec.Steps[i].CatalogTools = cataloglink.ResolveMentions(spec.Steps[i].ToolMentions)
	}
	if len(spec.CatalogTools) == 0 && len(sum.CatalogTools) > 0 {
		spec.CatalogTools = sum.CatalogTools
	}
	return spec, nil
}

// BySubdomain returns summaries for a subdomain.
func (c *Catalog) BySubdomain(subdomain string) []domain.ProcedureSummary {
	subdomain = strings.ToLower(strings.TrimSpace(subdomain))
	var out []domain.ProcedureSummary
	for _, p := range c.file.Procedures {
		if strings.EqualFold(p.Subdomain, subdomain) {
			out = append(out, p)
		}
	}
	return out
}
