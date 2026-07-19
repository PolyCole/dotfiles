package dots

import (
	"strings"
	"testing"
)

func TestCommitSubject(t *testing.T) {
	tests := []struct {
		name    string
		machine string
		paths   []string
		want    string
	}{
		{
			name:    "single category",
			machine: "personal",
			paths:   []string{"machines/personal/Brewfile"},
			want:    "🍺 personal: brew packages",
		},
		{
			name:    "several paths, one category",
			machine: "ibotta",
			paths:   []string{"modules/git.zsh", "shell/init.zsh", "zshrc"},
			want:    "🐚 ibotta: shell config",
		},
		{
			name:    "two categories get the mixed emoji",
			machine: "personal",
			paths:   []string{"machines/personal/Brewfile", "tools/dots/render.go"},
			want:    "🎲 personal: brew packages, tools",
		},
		{
			name:    "labels elide past the cap",
			machine: "personal",
			paths: []string{
				"machines/personal/Brewfile",
				"machines/personal/python_packages",
				"machines/personal/gitconfig",
				"tools/dots/render.go",
				"hooks/runner.zsh",
			},
			want: "🎲 personal: brew packages, package lists, git config + 2 more",
		},
		{
			name:    "singular elision",
			machine: "personal",
			paths: []string{
				"machines/personal/Brewfile",
				"machines/personal/python_packages",
				"machines/personal/gitconfig",
				"tools/dots/render.go",
			},
			want: "🎲 personal: brew packages, package lists, git config + 1 more",
		},
		{
			name:    "unrecognised path falls back",
			machine: "personal",
			paths:   []string{"some/random/file.txt"},
			want:    "🟣 personal: odds and ends",
		},
		{
			name:    "no paths",
			machine: "personal",
			paths:   nil,
			want:    "🟣 personal: changes",
		},
		{
			name:    "empty machine",
			machine: "",
			paths:   []string{"zshrc"},
			want:    "🐚 unknown: shell config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := commitSubject(tt.machine, tt.paths); got != tt.want {
				t.Errorf("commitSubject() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Subject order must not depend on the order git happens to list paths.
func TestCommitSubject_StableOrder(t *testing.T) {
	a := commitSubject("personal", []string{"tools/dots/render.go", "machines/personal/Brewfile"})
	b := commitSubject("personal", []string{"machines/personal/Brewfile", "tools/dots/render.go"})
	if a != b {
		t.Errorf("subject depends on path order: %q vs %q", a, b)
	}
}

// Narrower rules must win over broader ones that would also match.
func TestClassify_SpecificBeatsGeneral(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"machines/ibotta/Brewfile", "brew packages"},   // not "machine config"
		{"machines/ibotta/p10k.zsh", "prompt"},          // not "shell config"
		{"machines/personal/gitconfig", "git config"},   // not "machine config"
		{"machines/ibotta/aliases.zsh", "shell config"}, // .zsh before machines/
		{"hooks/pre-commit/lint.sh", "git hooks"},
		{"README.md", "docs"},
		{".beads/issues.jsonl", "issues"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			cats := classify([]string{tt.path})
			if len(cats) != 1 {
				t.Fatalf("classify(%q) returned %d categories, want 1", tt.path, len(cats))
			}
			if cats[0].label != tt.want {
				t.Errorf("classify(%q) = %q, want %q", tt.path, cats[0].label, tt.want)
			}
		})
	}
}

func TestParseStatusPaths(t *testing.T) {
	porcelain := strings.Join([]string{
		" M modules/git.zsh",
		"?? bin/dots",
		"A  machines/personal/Brewfile",
		"R  old/path.zsh -> new/path.zsh",
		"",
	}, "\n")

	got := parseStatusPaths(porcelain)
	want := []string{"modules/git.zsh", "bin/dots", "machines/personal/Brewfile", "new/path.zsh"}

	if len(got) != len(want) {
		t.Fatalf("parseStatusPaths() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("path %d = %q, want %q", i, got[i], want[i])
		}
	}
}
