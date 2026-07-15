package repository

import (
	"context"
	"github.com/butbeautifulv/veil/pkg/vuln/domain"
)

type VulnerabilityRepository interface {
	Save(ctx context.Context, v *domain.Vulnerability) error
	Upsert(ctx context.Context, v *domain.Vulnerability) error
	// MergeExploitForCVE links an Exploit node to an existing Vulnerability; no-op if CVE not in graph.
	MergeExploitForCVE(ctx context.Context, cve string, ref domain.ExploitRef) error
}
