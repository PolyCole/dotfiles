package dots

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempModule(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExtractDefinedNames(t *testing.T) {
	dir := t.TempDir()
	path := writeTempModule(t, dir, "m.zsh", `# modules/m.zsh
# Test module

# Commands:
#   foo  - does foo

alias foo="echo foo"
alias 1:1="echo hello"
function bar() {
  echo bar
}
baz() {
  echo baz
}
qux(){
  echo qux
}
# alias commented="nope"
`)

	names, err := extractDefinedNames(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"foo", "1:1", "bar", "baz", "qux"} {
		if !names[want] {
			t.Errorf("expected %q to be detected as defined", want)
		}
	}
	if names["commented"] {
		t.Error("commented-out alias should not be detected")
	}
}

func TestCheckModule(t *testing.T) {
	dir := t.TempDir()
	path := writeTempModule(t, dir, "drift.zsh", `# modules/drift.zsh
# Module with drift

# Commands:
#   real-cmd     - exists
#   ghost-cmd    - documented but not defined

real-cmd() {
  echo real
}
undocumented-cmd() {
  echo hidden
}
`)

	issues := checkModule(dir, path)

	var errors, warns int
	for _, issue := range issues {
		switch issue.severity {
		case "error":
			errors++
			if issue.message != `documented command "ghost-cmd" is not defined in this file` {
				t.Errorf("unexpected error message: %s", issue.message)
			}
		case "warn":
			warns++
		}
	}
	if errors != 1 {
		t.Errorf("expected 1 error (ghost-cmd), got %d: %v", errors, issues)
	}
	if warns != 1 {
		t.Errorf("expected 1 warn (undocumented-cmd), got %d: %v", warns, issues)
	}
}

func TestDocName(t *testing.T) {
	cases := map[string]string{
		"build_and_promote <service> <git-hash>": "build_and_promote",
		"weather [location]":                     "weather",
		"aoc":                                    "aoc",
		"":                                       "",
	}
	for in, want := range cases {
		if got := docName(in); got != want {
			t.Errorf("docName(%q) = %q, want %q", in, got, want)
		}
	}
}
