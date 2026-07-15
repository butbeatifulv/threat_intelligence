package domain

import (
	"testing"

	playbookdomain "github.com/butbeautifulv/veil/pkg/playbook/domain"
)

func TestKnowledgeAliasesCompile(t *testing.T) {
	var _ = SkillMeta(playbookdomain.SkillMeta{})
	var _ = ProcedureSpec(playbookdomain.ProcedureSpec{})
	if StepShell != playbookdomain.StepShell {
		t.Fatal("StepShell alias mismatch")
	}
}

func TestDefaultFrameworkContour(t *testing.T) {
	fc := DefaultFrameworkContour()
	if fc.MappingsDir != FrameworkMappingsRelDir {
		t.Fatalf("mappings dir: %q", fc.MappingsDir)
	}
	if fc.AttackNavigatorLayer != FrameworkAttackNavigatorRel {
		t.Fatalf("navigator layer: %q", fc.AttackNavigatorLayer)
	}
}
