package dots

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTemp creates a temporary .zsh file with the given content and returns
// its path. The caller is responsible for removing it (t.Cleanup handles this).
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.zsh")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeTemp: %v", err)
	}
	return path
}

// --- parseCommandEntry unit tests -------------------------------------------

func TestParseCommandEntry_Arrow(t *testing.T) {
	got := parseCommandEntry("  weather [location]   → Fetch weather from wttr.in")
	want := Command{Name: "weather [location]", Description: "Fetch weather from wttr.in"}
	if got != want {
		t.Errorf("arrow: got %+v, want %+v", got, want)
	}
}

func TestParseCommandEntry_Dash(t *testing.T) {
	got := parseCommandEntry("  git-purge-dir <dir>    - Remove directory from repo history")
	want := Command{Name: "git-purge-dir <dir>", Description: "Remove directory from repo history"}
	if got != want {
		t.Errorf("dash: got %+v, want %+v", got, want)
	}
}

func TestParseCommandEntry_DashNoHyphenAmbiguity(t *testing.T) {
	// Hyphens in the command name must not be mistaken for the separator.
	got := parseCommandEntry("  oncall-check    - Check production API server revisions")
	want := Command{Name: "oncall-check", Description: "Check production API server revisions"}
	if got != want {
		t.Errorf("dash hyphen: got %+v, want %+v", got, want)
	}
}

func TestParseCommandEntry_SpaceOnly(t *testing.T) {
	got := parseCommandEntry("  mycommand   Does something useful")
	want := Command{Name: "mycommand", Description: "Does something useful"}
	if got != want {
		t.Errorf("space-only: got %+v, want %+v", got, want)
	}
}

func TestParseCommandEntry_SpaceOnlyWithArgs(t *testing.T) {
	got := parseCommandEntry("  build_and_promote <service> <git-hash>   Build and promote a microservice")
	want := Command{Name: "build_and_promote <service> <git-hash>", Description: "Build and promote a microservice"}
	if got != want {
		t.Errorf("space-only with args: got %+v, want %+v", got, want)
	}
}

func TestParseCommandEntry_NameOnly(t *testing.T) {
	// A line with no separator at all is treated as a name with no description.
	got := parseCommandEntry("  mynewcommand")
	want := Command{Name: "mynewcommand"}
	if got != want {
		t.Errorf("name-only: got %+v, want %+v", got, want)
	}
}

func TestParseCommandEntry_Empty(t *testing.T) {
	got := parseCommandEntry("")
	if got.Name != "" || got.Description != "" {
		t.Errorf("empty: expected zero Command, got %+v", got)
	}
}

// --- parseFile / ParseFiles integration tests --------------------------------

func TestParseFile_DashFormat(t *testing.T) {
	content := `# modules/git.zsh
# Git utilities for history manipulation and commit management

# Commands:
#   git-purge-dir <dir>    - Remove directory from repo history (filter-branch)
#   git-purge-file <file>  - Remove file from repo history (filter-branch)
#   git-amend <message>    - Amend previous commit with new message

git-purge-dir() { }
`
	path := writeTemp(t, content)
	m, err := parseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.Name != "test" {
		t.Errorf("name: got %q, want %q", m.Name, "test")
	}
	if m.Description != "Git utilities for history manipulation and commit management" {
		t.Errorf("description: got %q", m.Description)
	}
	if len(m.Commands) != 3 {
		t.Fatalf("commands: got %d, want 3", len(m.Commands))
	}

	cases := []Command{
		{Name: "git-purge-dir <dir>", Description: "Remove directory from repo history (filter-branch)"},
		{Name: "git-purge-file <file>", Description: "Remove file from repo history (filter-branch)"},
		{Name: "git-amend <message>", Description: "Amend previous commit with new message"},
	}
	for i, want := range cases {
		if m.Commands[i] != want {
			t.Errorf("command[%d]: got %+v, want %+v", i, m.Commands[i], want)
		}
	}
}

func TestParseFile_ArrowFormat(t *testing.T) {
	content := `# modules/nav.zsh
# Navigation helpers

# Commands:
#   o [path]   → Open current or specified directory in Finder
#   a [path]   → Open current or specified directory in VS Code

o() { }
`
	path := writeTemp(t, content)
	m, err := parseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Commands) != 2 {
		t.Fatalf("commands: got %d, want 2", len(m.Commands))
	}
	want0 := Command{Name: "o [path]", Description: "Open current or specified directory in Finder"}
	if m.Commands[0] != want0 {
		t.Errorf("command[0]: got %+v, want %+v", m.Commands[0], want0)
	}
}

func TestParseFile_SpaceOnlyFormat(t *testing.T) {
	content := `# modules/tools.zsh
# Miscellaneous tools

# Commands:
#   build_and_promote <service> <git-hash>   Build and promote a microservice
#   reset_mono_stage <instance>              Tear down and recreate a staging instance

build_and_promote() { }
`
	path := writeTemp(t, content)
	m, err := parseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Commands) != 2 {
		t.Fatalf("commands: got %d, want 2", len(m.Commands))
	}
	want := Command{Name: "build_and_promote <service> <git-hash>", Description: "Build and promote a microservice"}
	if m.Commands[0] != want {
		t.Errorf("command[0]: got %+v, want %+v", m.Commands[0], want)
	}
}

func TestParseFile_BlockEndsAtNonComment(t *testing.T) {
	// Lines after the blank line (which is not a '#' line) must not be parsed
	// as commands even if they start with '#'.
	content := `# modules/example.zsh
# Example module

# Commands:
#   cmd-one   - Does one thing
#   cmd-two   - Does two things

# This comment is outside the block and must be ignored.
cmd-one() { }
`
	path := writeTemp(t, content)
	m, err := parseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Commands) != 2 {
		t.Fatalf("commands: got %d, want 2 (block must stop at blank line)", len(m.Commands))
	}
}

func TestParseFile_NoCommandsBlock(t *testing.T) {
	content := `# modules/simple.zsh
# A module with no Commands: block

alias foo="bar"
`
	path := writeTemp(t, content)
	m, err := parseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Commands) != 0 {
		t.Errorf("expected no commands, got %d", len(m.Commands))
	}
	if m.Description != "A module with no Commands: block" {
		t.Errorf("description: got %q", m.Description)
	}
}

func TestParseFile_ModuleName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kubernetes.zsh")
	content := "# modules/kubernetes.zsh\n# Kubernetes helpers\n\n# Commands:\n#   kubeconfig-prod - Switch to prod\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := parseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Name != "kubernetes" {
		t.Errorf("name: got %q, want %q", m.Name, "kubernetes")
	}
}

func TestParseFiles_MultipleFiles(t *testing.T) {
	git := writeTemp(t, "# modules/git.zsh\n# Git utils\n\n# Commands:\n#   git-amend - Amend commit\n")

	dir := t.TempDir()
	navPath := filepath.Join(dir, "navigation.zsh")
	if err := os.WriteFile(navPath, []byte("# modules/navigation.zsh\n# Nav utils\n\n# Commands:\n#   o [path] → Open in Finder\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	modules := ParseFiles([]string{git, navPath})
	if len(modules) != 2 {
		t.Fatalf("got %d modules, want 2", len(modules))
	}
}

func TestParseFiles_SkipsMissingFiles(t *testing.T) {
	modules := ParseFiles([]string{"/nonexistent/path/foo.zsh"})
	if len(modules) != 0 {
		t.Errorf("expected 0 modules for missing file, got %d", len(modules))
	}
}

// TestParseFile_RealModules exercises the parser against the actual module
// files found in this repository. It validates that every module loads without
// error and that known modules have expected properties.
func TestParseFile_RealModules(t *testing.T) {
	realModules := []struct {
		path        string
		wantName    string
		minCommands int
	}{
		{"../../modules/git.zsh", "git", 3},
		{"../../modules/aliases.zsh", "aliases", 4},
		{"../../modules/navigation.zsh", "navigation", 4},
		{"../../modules/web.zsh", "web", 2},
		{"../../modules/docker.zsh", "docker", 2},
		{"../../modules/crypto.zsh", "crypto", 3},
		{"../../modules/advent.zsh", "advent", 2},
		{"../../modules/hooks.zsh", "hooks", 1},
	}

	for _, tc := range realModules {
		t.Run(tc.wantName, func(t *testing.T) {
			absPath, err := filepath.Abs(tc.path)
			if err != nil {
				t.Skipf("could not resolve path %q: %v", tc.path, err)
			}
			m, err := parseFile(absPath)
			if err != nil {
				t.Fatalf("parseFile(%q): %v", absPath, err)
			}
			if m.Name != tc.wantName {
				t.Errorf("name: got %q, want %q", m.Name, tc.wantName)
			}
			if m.Description == "" {
				t.Errorf("description is empty")
			}
			if len(m.Commands) < tc.minCommands {
				t.Errorf("commands: got %d, want at least %d", len(m.Commands), tc.minCommands)
			}
		})
	}
}
