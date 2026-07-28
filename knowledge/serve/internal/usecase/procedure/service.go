package procedure

import (
	"github.com/butbeautifulv/veil/pkg/playbook/domain"
	pbprocedure "github.com/butbeautifulv/veil/pkg/playbook/procedure"
	pbframework "github.com/butbeautifulv/veil/pkg/playbook/framework"
)

// Service exposes structured procedures for HTTP/MCP.
type Service struct {
	cat *pbprocedure.Catalog
}

func NewService() (*Service, error) {
	cat, err := pbprocedure.Default()
	if err != nil {
		return nil, err
	}
	return &Service{cat: cat}, nil
}

func (s *Service) Catalog() *pbprocedure.Catalog { return s.cat }

func (s *Service) GetSpec(id string) (domain.ProcedureSpec, error) {
	return s.cat.GetSpec(id)
}

func (s *Service) RecommendTools(id string) ([]string, error) {
	sum, ok := s.cat.GetSummary(id)
	if !ok {
		spec, err := s.cat.GetSpec(id)
		if err != nil {
			return nil, err
		}
		return spec.CatalogTools, nil
	}
	if len(sum.CatalogTools) > 0 {
		return sum.CatalogTools, nil
	}
	spec, err := s.cat.GetSpec(id)
	if err != nil {
		return nil, err
	}
	return spec.CatalogTools, nil
}

func (s *Service) Subdomains() ([]pbframework.SubdomainEntry, error) {
	return pbframework.LoadSubdomains()
}

func (s *Service) TechniqueSkillIDs(tid string) ([]string, error) {
	return pbframework.SkillsForTechnique(tid)
}

func (s *Service) CatalogToolsForTechnique(tid string) []string {
	return pbprocedure.CatalogToolsForTechnique(tid, s.cat.Meta().Procedures)
}
