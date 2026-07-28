package procedure

import (
	"strings"

	"github.com/butbeautifulv/veil/pkg/playbook/domain"
)

// CatalogToolsForTechnique returns catalog tool ids referenced by skills for a MITRE technique.
func CatalogToolsForTechnique(techniqueID string, procedures []domain.ProcedureSummary) []string {
	tid := strings.ToUpper(strings.TrimSpace(techniqueID))
	seen := map[string]struct{}{}
	var out []string
	for _, p := range procedures {
		for _, a := range p.AttackIDs {
			if strings.EqualFold(a, tid) {
				for _, t := range p.CatalogTools {
					if _, dup := seen[t]; dup {
						continue
					}
					seen[t] = struct{}{}
					out = append(out, t)
				}
				break
			}
		}
	}
	return out
}
