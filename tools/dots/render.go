package dots

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	colorGroupName  = lipgloss.Color("#7DCFFF") // bright cyan
	colorCmdName    = lipgloss.Color("#9ECE6A") // green
	colorArgBracket = lipgloss.Color("#E0AF68") // amber/orange for [optional]
	colorArgAngle   = lipgloss.Color("#FF9E64") // orange for <required>
	colorDesc       = lipgloss.Color("#A9B1D6") // muted blue-gray
	colorDim        = lipgloss.Color("#565F89") // dim purple-gray
	colorHighlight  = lipgloss.Color("#F7768E") // pink/red for search matches
	colorCount      = lipgloss.Color("#BB9AF7") // lavender for counts
	colorBorder     = lipgloss.Color("#3B4261") // dark border
)

// Styles
var (
	styleGroupHeader = lipgloss.NewStyle().
				Foreground(colorGroupName).
				Bold(true)

	styleGroupBorder = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBorder).
				Padding(0, 1)

	styleCmdName = lipgloss.NewStyle().
			Foreground(colorCmdName).
			Bold(true)

	styleDesc = lipgloss.NewStyle().
			Foreground(colorDesc)

	styleDim = lipgloss.NewStyle().
			Foreground(colorDim)

	styleCount = lipgloss.NewStyle().
			Foreground(colorCount)

	styleHighlight = lipgloss.NewStyle().
			Foreground(colorHighlight).
			Bold(true)

	styleHint = lipgloss.NewStyle().
			Foreground(colorDim).
			Italic(true)
)

// reArgBracket matches [optional-arg] patterns
var reArgBracket = regexp.MustCompile(`\[[^\]]+\]`)

// reArgAngle matches <required-arg> patterns
var reArgAngle = regexp.MustCompile(`<[^>]+>`)

// renderCmdName applies color to a command name, coloring [brackets] and <angles> differently.
func renderCmdName(name string) string {
	// We'll build the result by finding bracket/angle spans and coloring them.
	type span struct {
		start, end int
		kind       string // "bracket" or "angle"
	}

	var spans []span
	for _, loc := range reArgBracket.FindAllStringIndex(name, -1) {
		spans = append(spans, span{loc[0], loc[1], "bracket"})
	}
	for _, loc := range reArgAngle.FindAllStringIndex(name, -1) {
		spans = append(spans, span{loc[0], loc[1], "angle"})
	}

	// Sort spans by start position
	for i := 0; i < len(spans); i++ {
		for j := i + 1; j < len(spans); j++ {
			if spans[j].start < spans[i].start {
				spans[i], spans[j] = spans[j], spans[i]
			}
		}
	}

	if len(spans) == 0 {
		return styleCmdName.Render(name)
	}

	var sb strings.Builder
	pos := 0
	for _, sp := range spans {
		// Text before this span: render as cmd name color
		if sp.start > pos {
			sb.WriteString(styleCmdName.Render(name[pos:sp.start]))
		}
		// The arg span
		argText := name[sp.start:sp.end]
		if sp.kind == "bracket" {
			sb.WriteString(lipgloss.NewStyle().Foreground(colorArgBracket).Render(argText))
		} else {
			sb.WriteString(lipgloss.NewStyle().Foreground(colorArgAngle).Render(argText))
		}
		pos = sp.end
	}
	// Remaining text
	if pos < len(name) {
		sb.WriteString(styleCmdName.Render(name[pos:]))
	}
	return sb.String()
}

const banner = ` ██████████         ███████       ███████████     █████████
░░███░░░░███      ███░░░░░███    ░█░░░███░░░█    ███░░░░░███
 ░███   ░░███    ███     ░░███   ░   ░███  ░    ░███    ░░░
 ░███    ░███   ░███      ░███       ░███       ░░█████████
 ░███    ░███   ░███      ░███       ░███        ░░░░░░░░███
 ░███    ███    ░░███     ███        ░███        ███    ░███
 ██████████   ██ ░░░███████░   ██    █████    ██░░█████████
░░░░░░░░░░   ░░    ░░░░░░░    ░░    ░░░░░    ░░  ░░░░░░░░░  `

// RenderOverview writes the overview (dots with no args) to w.
func RenderOverview(w io.Writer, modules []Module) {
	// Banner
	bannerStyle := lipgloss.NewStyle().Foreground(colorGroupName)
	fmt.Fprintf(w, "%s\n", bannerStyle.Render(banner))
	subtitle := styleDim.Render("Don't Overthink This Shit")
	fmt.Fprintf(w, "%s\n\n", subtitle)

	// Find longest group name for alignment
	maxLen := 0
	for _, m := range modules {
		if len(m.Name) > maxLen {
			maxLen = len(m.Name)
		}
	}

	// Table of groups
	for _, m := range modules {
		count := len(m.Commands)
		groupName := styleGroupHeader.Render(fmt.Sprintf("%-*s", maxLen, m.Name))
		countStr := styleCount.Render(fmt.Sprintf("%2d", count))
		plural := "commands"
		if count == 1 {
			plural = "command"
		}
		fmt.Fprintf(w, "  %s  %s %s\n", groupName, countStr, styleDim.Render(plural))
	}

	// Hints
	fmt.Fprintf(w, "\n")
	hints := []string{
		"dots <group>          — detail for one group",
		"dots --all            — every command across all groups",
		"dots --search <term>  — search command descriptions",
	}
	for _, h := range hints {
		fmt.Fprintf(w, "  %s\n", styleHint.Render(h))
	}
}

// RenderGroup writes the detail view for a single group.
func RenderGroup(w io.Writer, m Module) {
	header := renderGroupHeader(m)
	fmt.Fprintln(w, header)
	fmt.Fprintln(w)
	renderCommands(w, m.Commands, "")
	fmt.Fprintln(w)
}

// RenderAll writes all groups with section headers.
func RenderAll(w io.Writer, modules []Module) {
	for i, m := range modules {
		if len(m.Commands) == 0 {
			continue
		}
		if i > 0 {
			fmt.Fprintln(w)
		}
		header := renderGroupHeader(m)
		fmt.Fprintln(w, header)
		fmt.Fprintln(w)
		renderCommands(w, m.Commands, "")
	}
}

// RenderSearch writes filtered results with match highlighting.
func RenderSearch(w io.Writer, modules []Module, term string) {
	if term == "" {
		fmt.Fprintln(w, styleDesc.Render("Usage: dots --search <term>"))
		return
	}
	lower := strings.ToLower(term)
	found := false

	for _, m := range modules {
		var matched []Command
		for _, cmd := range m.Commands {
			if strings.Contains(strings.ToLower(cmd.Name), lower) ||
				strings.Contains(strings.ToLower(cmd.Description), lower) {
				matched = append(matched, cmd)
			}
		}
		if len(matched) == 0 {
			continue
		}
		if found {
			fmt.Fprintln(w)
		}
		found = true
		header := renderGroupHeader(m)
		fmt.Fprintln(w, header)
		fmt.Fprintln(w)
		renderCommands(w, matched, term)
	}

	if !found {
		msg := fmt.Sprintf("No commands found matching %q.", term)
		fmt.Fprintln(w, styleDim.Render(msg))
	}
}

// renderGroupHeader renders a styled group header line with border.
func renderGroupHeader(m Module) string {
	name := styleGroupHeader.Render(m.Name)
	if m.Description != "" {
		desc := styleDim.Render("— " + m.Description)
		return styleGroupBorder.Render(name + "  " + desc)
	}
	return styleGroupBorder.Render(name)
}

// renderCommands renders a command list to w. If highlight is non-empty, matching
// substrings in command names and descriptions are highlighted.
func renderCommands(w io.Writer, commands []Command, highlight string) {
	// Find max command name length (visual, not accounting for ANSI codes
	// since we measure the raw string before styling).
	maxLen := 0
	for _, cmd := range commands {
		if len(cmd.Name) > maxLen {
			maxLen = len(cmd.Name)
		}
	}

	for _, cmd := range commands {
		// Pad name to align descriptions
		padding := maxLen - len(cmd.Name)
		nameRendered := renderCmdName(cmd.Name) + strings.Repeat(" ", padding)

		var descRendered string
		if cmd.Description != "" {
			descText := cmd.Description
			if highlight != "" {
				descText = highlightTerm(descText, highlight)
				nameRendered = highlightTermInRendered(nameRendered, cmd.Name, highlight)
			}
			descRendered = styleDesc.Render("  " + descText)
		}

		fmt.Fprintf(w, "  %s%s\n", nameRendered, descRendered)
	}
}

// highlightTerm wraps occurrences of term in s with the highlight style (case-insensitive).
func highlightTerm(s, term string) string {
	if term == "" {
		return s
	}
	lower := strings.ToLower(s)
	lowerTerm := strings.ToLower(term)
	var result strings.Builder
	pos := 0
	for {
		idx := strings.Index(lower[pos:], lowerTerm)
		if idx == -1 {
			result.WriteString(s[pos:])
			break
		}
		abs := pos + idx
		result.WriteString(s[pos:abs])
		result.WriteString(styleHighlight.Render(s[abs : abs+len(term)]))
		pos = abs + len(term)
	}
	return result.String()
}

// highlightTermInRendered re-renders the command name with highlight applied to
// the raw name text. We need to re-do the full render to preserve argument coloring.
func highlightTermInRendered(alreadyRendered, rawName, term string) string {
	// If term appears in rawName, we rebuild the rendered name with highlight.
	if !strings.Contains(strings.ToLower(rawName), strings.ToLower(term)) {
		return alreadyRendered
	}
	// Re-render with highlight embedded — simplest approach: highlight in the
	// raw name first, then render base color around it.
	highlighted := highlightTerm(rawName, term)
	// We can't easily compose highlight + arg coloring without complex parsing,
	// so for the command name in search results we just return the highlighted version
	// without arg coloring (the highlight is more important for discoverability).
	return highlighted
}
