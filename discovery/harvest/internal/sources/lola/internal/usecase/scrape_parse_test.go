package usecase

import (
	"strings"
	"testing"

	"github.com/butbeautifulv/veil/pkg/lola/domain"
)

func TestParseLOLBASYAML(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		raw     string
		wantErr bool
		check   func(t *testing.T, a *domain.Artifact)
	}{
		{
			name: "minimal name only",
			raw:  "Name: cmd.exe\n",
			check: func(t *testing.T, a *domain.Artifact) {
				t.Helper()
				if a.Name != "cmd.exe" {
					t.Fatalf("Name=%q", a.Name)
				}
			},
		},
		{
			name: "full fields",
			raw: `Name: rundll32.exe
Description: Loads DLLs
Author: Alice
OS: Windows
MITRE:
  ID: T1218.011
Categories:
  - Execute
Full Path:
  - C:\Windows\System32\rundll32.exe
Commands:
  - Command: rundll32.exe shell32.dll,Control_RunDLL
    Description: Run DLL
    Usecase: Proxy execution
Detection:
  Sigma:
    - https://example.com/sigma.yml
  YARA:
    - rule_placeholder
`,
			check: func(t *testing.T, a *domain.Artifact) {
				t.Helper()
				if a.Name != "rundll32.exe" {
					t.Fatalf("Name=%q", a.Name)
				}
				if !strings.Contains(a.Description, "Loads DLLs") || !strings.Contains(a.Description, "Author: Alice") {
					t.Fatalf("Description=%q", a.Description)
				}
				if len(a.OS) != 1 || a.OS[0] != "Windows" {
					t.Fatalf("OS=%v", a.OS)
				}
				if a.MitreID != "T1218.011" {
					t.Fatalf("MitreID=%q", a.MitreID)
				}
				if a.Category != "Execute" {
					t.Fatalf("Category=%q", a.Category)
				}
				if len(a.Paths) != 1 {
					t.Fatalf("Paths=%v", a.Paths)
				}
				if len(a.Commands) != 1 || a.Commands[0].Usecase != "Proxy execution" {
					t.Fatalf("Commands=%v", a.Commands)
				}
				if len(a.Detection.Sigma) != 1 || len(a.Detection.Yara) != 1 {
					t.Fatalf("Detection=%+v", a.Detection)
				}
			},
		},
		{
			name:    "missing name",
			raw:     "Description: orphan\n",
			wantErr: true,
		},
		{
			name:    "invalid yaml",
			raw:     "Name: [\n",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, err := parseLOLBASYAML([]byte(tt.raw))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.check != nil {
				tt.check(t, a)
			}
		})
	}
}

func TestParseGTFOBinsMarkdown(t *testing.T) {
	t.Parallel()
	body := "# sudo\n\nRun as root.\n\n## Shell\n\necho pwn"
	a := parseGTFOBinsMarkdown("sudo.md", body)
	if a.Name != "sudo" {
		t.Fatalf("Name=%q", a.Name)
	}
	if a.Category != "gtfobins" {
		t.Fatalf("Category=%q", a.Category)
	}
	if len(a.OS) != 1 || a.OS[0] != "linux" {
		t.Fatalf("OS=%v", a.OS)
	}
	if !strings.Contains(a.Description, "Run as root") {
		t.Fatalf("Description=%q", a.Description)
	}
	if strings.Contains(a.Description, "## Shell") {
		t.Fatalf("section should be trimmed from description, got %q", a.Description)
	}
}

func TestParseGTFOBinsMarkdown_underscoreName(t *testing.T) {
	t.Parallel()
	a := parseGTFOBinsMarkdown("bin_bash.md", "intro")
	if a.Name != "bin/bash" {
		t.Fatalf("Name=%q", a.Name)
	}
}

func TestFirstParagraph(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "hello world", "hello world"},
		{"stops at h2", "# title\n\nfirst para\n\n## next\n\nrest", "# title\n\nfirst para"},
		{"truncates long", strings.Repeat("x", 5000), ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := firstParagraph(tt.in)
			if tt.name == "truncates long" {
				if len(got) <= 4000 || !strings.HasSuffix(got, "…") {
					t.Fatalf("len=%d got suffix=%v", len(got), strings.HasSuffix(got, "…"))
				}
				return
			}
			if got != tt.want {
				t.Fatalf("got %q", got)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()
	if got := truncate("short", 10); got != "short" {
		t.Fatalf("got %q", got)
	}
	long := strings.Repeat("a", 20)
	got := truncate(long, 10)
	if len(got) <= 10 || !strings.HasSuffix(got, "…") {
		t.Fatalf("got len=%d %q", len(got), got)
	}
}
