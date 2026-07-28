package domain

import (
	coderulesdomain "github.com/butbeautifulv/veil/pkg/coderules/domain"
	dsdomain "github.com/butbeautifulv/veil/pkg/ds/domain"
	nucleidomain "github.com/butbeautifulv/veil/pkg/nuclei/domain"
	sbomdomain "github.com/butbeautifulv/veil/pkg/sbom/domain"
)

// SourceRefFromAdvisoryRef maps an SBOM advisory thin ref to SourceRef.
func SourceRefFromAdvisoryRef(a sbomdomain.AdvisoryRef) SourceRef {
	src := Source(a.Source)
	if src == "" {
		src = SourceSBOM
	}
	return SourceRef{
		Source: src,
		Key:    a.CVE,
		Path:   a.Path,
	}
}

// SourceRefFromTemplate maps a Nuclei template identity to SourceRef.
func SourceRefFromTemplate(t nucleidomain.Template) SourceRef {
	key := t.ID
	if key == "" {
		key = t.Path
	}
	return SourceRef{
		Source: SourceNuclei,
		Key:    key,
		Path:   t.Path,
	}
}

// SourceRefFromResource maps a detections/signals resource to SourceRef.
func SourceRefFromResource(r dsdomain.Resource) SourceRef {
	src := Source(r.Source)
	if src == "" {
		src = SourceDS
	}
	return SourceRef{
		Source: src,
		Key:    r.Key,
		Path:   r.URL,
		Kind:   r.Kind,
	}
}

// SourceRefFromRuleFile maps a code-rules file identity to SourceRef.
func SourceRefFromRuleFile(f coderulesdomain.RuleFile) SourceRef {
	key := f.Path
	if f.Repo != "" && f.Path != "" {
		key = f.Repo + "/" + f.Path
	} else if f.Repo != "" {
		key = f.Repo
	}
	return SourceRef{
		Source: SourceCoderules,
		Key:    key,
		Path:   f.Path,
		Kind:   f.Format,
	}
}
