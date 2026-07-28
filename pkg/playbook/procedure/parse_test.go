package procedure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/butbeautifulv/veil/pkg/playbook/domain"
	pbindex "github.com/butbeautifulv/veil/pkg/playbook/index"
)

func TestParseSkillMarkdown_diskImaging(t *testing.T) {
	root, err := pbindex.RepoRoot()
	if err != nil {
		t.Skip(err)
	}
	path := filepath.Join(root, "corpus/anthropic-cybersecurity-skills/skills/acquiring-disk-image-with-dd-and-dcfldd/SKILL.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skip(err)
	}
	spec := ParseSkillMarkdown("acquiring-disk-image-with-dd-and-dcfldd", "digital-forensics", nil, nil, string(raw))
	if len(spec.WhenToUse) < 2 {
		t.Fatalf("when_to_use: %d", len(spec.WhenToUse))
	}
	if len(spec.Steps) < 2 {
		t.Fatalf("steps: %d", len(spec.Steps))
	}
}

func TestParseSkillMarkdown_fixtureInline(t *testing.T) {
	raw := `---
name: inline
---
## When to use
- Case A
- Case B

## Prerequisites
- Root access

## Workflow
### Step 1: Scan
Use nmap and NMAP on hosts.

### Step 2: Shell step
Run echo hi in bash.

## Tools
masscan ffuf

## Scenarios
- Lab only
`
	spec := ParseSkillMarkdown("inline", "penetration-testing", []string{"T1046"}, nil, raw)
	if len(spec.WhenToUse) != 2 || len(spec.Prerequisites) != 1 {
		t.Fatalf("bullets: when=%d prereq=%d", len(spec.WhenToUse), len(spec.Prerequisites))
	}
	if len(spec.Scenarios) != 1 {
		t.Fatalf("scenarios: %d", len(spec.Scenarios))
	}
	if len(spec.Steps) != 2 {
		t.Fatalf("steps: %d", len(spec.Steps))
	}
	if spec.Steps[0].Kind != domain.StepTool {
		t.Fatalf("step1 kind: %s", spec.Steps[0].Kind)
	}
	if spec.Steps[1].Kind != domain.StepManual {
		t.Fatalf("step2 kind: %s", spec.Steps[1].Kind)
	}
	if len(spec.ToolMentions) < 3 {
		t.Fatalf("mentions: %v", spec.ToolMentions)
	}
}

func TestSplitSections(t *testing.T) {
	body := "## Workflow\nline1\n\n## Tools\nnmap\n\n## When to use\n- a\n"
	secs := splitSections(body)
	if secs["workflow"] != "line1" {
		t.Fatalf("workflow: %q", secs["workflow"])
	}
	if secs["tools"] != "nmap" {
		t.Fatalf("tools: %q", secs["tools"])
	}
	if secs["when to use"] != "- a" {
		t.Fatalf("when: %q", secs["when to use"])
	}
}

func TestExtractSteps(t *testing.T) {
	workflow := `### Step 1: Alpha
Run nmap here.

### Step 2: Beta
Manual only.
`
	steps := extractSteps(workflow)
	if len(steps) != 2 {
		t.Fatalf("steps: %d", len(steps))
	}
	if steps[0].Number != 1 || steps[0].Title != "Alpha" {
		t.Fatalf("step1: %+v", steps[0])
	}
	if steps[0].Kind != domain.StepTool {
		t.Fatalf("step1 kind: %s", steps[0].Kind)
	}
	if steps[1].Kind != domain.StepManual {
		t.Fatalf("step2 kind: %s", steps[1].Kind)
	}
	if len(extractSteps("no steps here")) != 0 {
		t.Fatal("expected no steps")
	}
	shell := extractSteps("### Step 1: Run\n```bash\necho hi\n```\n")
	if len(shell) != 1 || shell[0].Kind != domain.StepShell {
		t.Fatalf("shell step: %+v", shell)
	}
}

func TestTokenizeTools(t *testing.T) {
	got := tokenizeTools("nmap NMAP and nuclei; nuclei again")
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
	if got[0] != "nmap" || got[1] != "nuclei" {
		t.Fatalf("tokens: %v", got)
	}
	if len(tokenizeTools("no tools listed")) != 0 {
		t.Fatal("expected empty")
	}
}

func TestCollectMentions(t *testing.T) {
	steps := []domain.ProcedureStep{
		{ToolMentions: []string{"nmap", "nuclei"}},
		{ToolMentions: []string{"nmap"}},
	}
	got := collectMentions(steps, "masscan")
	if len(got) != 3 {
		t.Fatalf("got %v", got)
	}
}
