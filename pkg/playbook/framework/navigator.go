// Package framework loads committed security framework mappings (MITRE Navigator layer, etc.).
package framework

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	pbindex "github.com/butbeautifulv/veil/pkg/playbook/index"
)

// NavigatorLayer is a subset of MITRE ATT&CK Navigator layer JSON (v4.x).
type NavigatorLayer struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Domain      string              `json:"domain"`
	Versions    map[string]string   `json:"versions"`
	Techniques  []NavigatorTechnique `json:"techniques"`
}

// NavigatorTechnique is one technique entry in the layer.
type NavigatorTechnique struct {
	TechniqueID string  `json:"techniqueID"`
	Score       float64 `json:"score"`
	Comment     string  `json:"comment"`
}

var frameworkMappingsDir = pbindex.MappingsDir

// LoadNavigatorLayer reads attack-navigator-layer.json from the committed mappings dir.
func LoadNavigatorLayer() (*NavigatorLayer, error) {
	dir, err := frameworkMappingsDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "attack-navigator-layer.json")
	raw, err := os.ReadFile(path) // #nosec G304 -- dir is the discovered committed mappings dir; filename is a literal
	if err != nil {
		return nil, fmt.Errorf("framework: read %s: %w", path, err)
	}
	var layer NavigatorLayer
	if err := json.Unmarshal(raw, &layer); err != nil {
		return nil, fmt.Errorf("framework: decode navigator layer: %w", err)
	}
	return &layer, nil
}

// TechniqueScores returns techniqueID -> score for techniques present in the layer.
func (l *NavigatorLayer) TechniqueScores() map[string]float64 {
	out := make(map[string]float64, len(l.Techniques))
	for _, t := range l.Techniques {
		if t.TechniqueID != "" {
			out[t.TechniqueID] = t.Score
		}
	}
	return out
}

// CoverageSummary is a compact view for API responses.
type CoverageSummary struct {
	LayerName       string  `json:"layer_name"`
	TechniqueCount  int     `json:"technique_count"`
	Domain          string  `json:"domain"`
	AttackVersion   string  `json:"attack_version,omitempty"`
}

// Summarize builds API-friendly metadata from the layer.
func (l *NavigatorLayer) Summarize() CoverageSummary {
	ver := ""
	if l.Versions != nil {
		ver = l.Versions["attack"]
	}
	return CoverageSummary{
		LayerName:      l.Name,
		TechniqueCount: len(l.Techniques),
		Domain:         l.Domain,
		AttackVersion:  ver,
	}
}
