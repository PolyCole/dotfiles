package dots

import (
	"strings"
	"testing"
)

func TestRenderOverview_ContainsGroupNames(t *testing.T) {
	modules := []Module{
		{Name: "git", Description: "Git utils", Commands: []Command{
			{Name: "git-amend", Description: "Amend commit"},
			{Name: "git-purge", Description: "Purge history"},
		}},
		{Name: "navigation", Description: "Nav helpers", Commands: []Command{
			{Name: "o", Description: "Open Finder"},
		}},
	}

	var sb strings.Builder
	RenderOverview(&sb, modules)
	out := sb.String()

	if !strings.Contains(out, "git") {
		t.Error("overview should contain group name 'git'")
	}
	if !strings.Contains(out, "navigation") {
		t.Error("overview should contain group name 'navigation'")
	}
	if !strings.Contains(out, "2") {
		t.Error("overview should show command count 2 for git")
	}
	if !strings.Contains(out, "D.O.T.S.") {
		t.Error("overview should contain title")
	}
}

func TestRenderGroup_ContainsCommands(t *testing.T) {
	m := Module{
		Name:        "web",
		Description: "Web utilities",
		Commands: []Command{
			{Name: "weather [location]", Description: "Fetch weather"},
			{Name: "status-dog <code>", Description: "HTTP status dog"},
		},
	}

	var sb strings.Builder
	RenderGroup(&sb, m)
	out := sb.String()

	if !strings.Contains(out, "weather") {
		t.Error("group output should contain command name 'weather'")
	}
	if !strings.Contains(out, "Fetch weather") {
		t.Error("group output should contain description")
	}
	if !strings.Contains(out, "web") {
		t.Error("group output should contain group name")
	}
}

func TestRenderAll_AllGroupsPresent(t *testing.T) {
	modules := []Module{
		{Name: "git", Commands: []Command{{Name: "git-amend", Description: "Amend"}}},
		{Name: "docker", Commands: []Command{{Name: "murderdocker", Description: "Prune all"}}},
	}

	var sb strings.Builder
	RenderAll(&sb, modules)
	out := sb.String()

	if !strings.Contains(out, "git") {
		t.Error("--all output should contain group 'git'")
	}
	if !strings.Contains(out, "docker") {
		t.Error("--all output should contain group 'docker'")
	}
	if !strings.Contains(out, "git-amend") {
		t.Error("--all output should contain command 'git-amend'")
	}
}

func TestRenderAll_SkipsEmptyGroups(t *testing.T) {
	modules := []Module{
		{Name: "empty", Commands: nil},
		{Name: "git", Commands: []Command{{Name: "git-amend", Description: "Amend"}}},
	}

	var sb strings.Builder
	RenderAll(&sb, modules)
	out := sb.String()

	if strings.Contains(out, "empty") {
		t.Error("--all should skip groups with no commands")
	}
}

func TestRenderSearch_FindsByName(t *testing.T) {
	modules := []Module{
		{Name: "git", Commands: []Command{
			{Name: "git-amend <message>", Description: "Amend commit"},
		}},
		{Name: "docker", Commands: []Command{
			{Name: "murderdocker", Description: "Prune all resources"},
		}},
	}

	var sb strings.Builder
	RenderSearch(&sb, modules, "amend")
	out := sb.String()

	if !strings.Contains(out, "git") {
		t.Error("search for 'amend' should find git group")
	}
	if strings.Contains(out, "murderdocker") {
		t.Error("search for 'amend' should not return murderdocker")
	}
}

func TestRenderSearch_FindsByDescription(t *testing.T) {
	modules := []Module{
		{Name: "docker", Commands: []Command{
			{Name: "murderdocker", Description: "Aggressively prune all Docker resources"},
		}},
	}

	var sb strings.Builder
	RenderSearch(&sb, modules, "prune")
	out := sb.String()

	if !strings.Contains(out, "murderdocker") {
		t.Error("search for 'prune' should find murderdocker via description")
	}
}

func TestRenderSearch_CaseInsensitive(t *testing.T) {
	modules := []Module{
		{Name: "git", Commands: []Command{
			{Name: "git-amend", Description: "Amend commit message"},
		}},
	}

	var sb strings.Builder
	RenderSearch(&sb, modules, "AMEND")
	out := sb.String()

	if !strings.Contains(out, "git") {
		t.Error("search should be case-insensitive")
	}
}

func TestRenderSearch_NotFound(t *testing.T) {
	modules := []Module{
		{Name: "git", Commands: []Command{
			{Name: "git-amend", Description: "Amend commit"},
		}},
	}

	var sb strings.Builder
	RenderSearch(&sb, modules, "xyznotfound")
	out := sb.String()

	if !strings.Contains(out, "No commands found") {
		t.Error("search with no results should print a 'not found' message")
	}
}

func TestRenderSearch_EmptyTerm(t *testing.T) {
	var sb strings.Builder
	RenderSearch(&sb, nil, "")
	out := sb.String()

	if !strings.Contains(out, "Usage") {
		t.Error("empty search term should print usage hint")
	}
}

func TestRenderCmdName_PlainName(t *testing.T) {
	// Plain names (no args) should render without crashing and contain the name text.
	out := renderCmdName("git-amend")
	if !strings.Contains(out, "git-amend") {
		t.Errorf("renderCmdName: expected 'git-amend' in output, got %q", out)
	}
}

func TestRenderCmdName_WithArgs(t *testing.T) {
	out := renderCmdName("weather [location]")
	if !strings.Contains(out, "weather") {
		t.Errorf("renderCmdName: expected 'weather' in output")
	}
	if !strings.Contains(out, "[location]") {
		t.Errorf("renderCmdName: expected '[location]' in output")
	}
}

func TestHighlightTerm(t *testing.T) {
	result := highlightTerm("Amend commit message", "amend")
	// The result should still contain the original text (possibly with ANSI codes)
	if !strings.Contains(strings.ToLower(result), "amend") {
		t.Error("highlightTerm should preserve the matched text")
	}
}
