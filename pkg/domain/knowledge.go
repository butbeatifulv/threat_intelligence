package domain

import playbookdomain "github.com/butbeautifulv/veil/pkg/playbook/domain"

// Playbook / knowledge contour type aliases (SOT remains pkg/playbook/domain).
type (
	SkillMeta           = playbookdomain.SkillMeta
	SkillDetail         = playbookdomain.SkillDetail
	IndexFile           = playbookdomain.IndexFile
	ProcedureSpec       = playbookdomain.ProcedureSpec
	ProcedureSummary    = playbookdomain.ProcedureSummary
	ProcedureStep       = playbookdomain.ProcedureStep
	ProceduresIndexFile = playbookdomain.ProceduresIndexFile
	StepKind            = playbookdomain.StepKind
)

// Step kind re-exports.
const (
	StepShell  = playbookdomain.StepShell
	StepManual = playbookdomain.StepManual
	StepTool   = playbookdomain.StepTool
)

// Framework mapping artifact paths (metadata only; JSON bodies stay on disk).
const (
	FrameworkMappingsRelDir     = "pkg/playbook/corpus/mappings"
	FrameworkCorpusVersionRel   = "pkg/playbook/corpus/VERSION"
	FrameworkAttackNavigatorRel = "pkg/playbook/corpus/mappings/attack-navigator-layer.json"
)

// FrameworkContour describes committed framework mapping artifacts (not loaded JSON).
type FrameworkContour struct {
	MappingsDir          string
	CorpusVersionFile    string
	AttackNavigatorLayer string
	NISTCSFDir           string
	OWASPDir             string
}

// DefaultFrameworkContour returns stable relative paths for playbook corpus mappings.
func DefaultFrameworkContour() FrameworkContour {
	return FrameworkContour{
		MappingsDir:          FrameworkMappingsRelDir,
		CorpusVersionFile:    FrameworkCorpusVersionRel,
		AttackNavigatorLayer: FrameworkAttackNavigatorRel,
		NISTCSFDir:           FrameworkMappingsRelDir + "/nist-csf",
		OWASPDir:             FrameworkMappingsRelDir + "/owasp",
	}
}
