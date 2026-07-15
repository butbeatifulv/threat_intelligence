package framework

import (
	"os"
	"path/filepath"

	pbframework "github.com/butbeautifulv/veil/pkg/playbook/framework"
	pbindex "github.com/butbeautifulv/veil/pkg/playbook/index"
)

// Service exposes committed framework mappings for HTTP/MCP.
type Service struct{}

func NewService() *Service { return &Service{} }

// MitreLayer returns parsed Navigator layer from pkg/playbook/corpus/mappings.
func (s *Service) MitreLayer() (*pbframework.NavigatorLayer, error) {
	return pbframework.LoadNavigatorLayer()
}

// MitreCoverage returns summary stats for the Navigator layer.
func (s *Service) MitreCoverage() (map[string]any, error) {
	layer, err := s.MitreLayer()
	if err != nil {
		return nil, err
	}
	sum := layer.Summarize()
	return map[string]any{
		"summary":    sum,
		"techniques": len(layer.Techniques),
	}, nil
}

// RawMitreLayerJSON reads the committed layer file bytes.
func (s *Service) RawMitreLayerJSON() ([]byte, error) {
	dir, err := pbindex.MappingsDir()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(dir, "attack-navigator-layer.json")) // #nosec G304 -- dir is the discovered committed mappings dir; filename is a literal
}

// ListMappingDocs returns relative paths to committed mapping markdown/json.
func (s *Service) ListMappingDocs() ([]string, error) {
	dir, err := pbindex.MappingsDir()
	if err != nil {
		return nil, err
	}
	var out []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		out = append(out, rel)
		return nil
	})
	return out, nil
}
