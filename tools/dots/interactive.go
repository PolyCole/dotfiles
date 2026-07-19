package dots

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// viewMode controls which screen is shown.
type viewMode int

const (
	modeOverview viewMode = iota
	modeDetail
	modeSearch
)

// interactiveModel is the bubbletea model for interactive dots.
type interactiveModel struct {
	modules  []Module
	mode     viewMode
	cursor   int    // selected group index in overview/search
	detail   Module // group shown in detail mode
	search   textinput.Model
	filtered []Module // groups matching current search query
	width    int
	height   int
}

// Interactive styles (re-use colors from render.go)
var (
	styleSelected = lipgloss.NewStyle().
			Foreground(colorGroupName).
			Bold(true).
			Background(lipgloss.Color("#1A1B2E"))

	styleSearchPrompt = lipgloss.NewStyle().
				Foreground(colorHighlight).
				Bold(true)

	styleStatusBar = lipgloss.NewStyle().
			Foreground(colorDim).
			Italic(true)

	styleDetailHeader = lipgloss.NewStyle().
				Foreground(colorGroupName).
				Bold(true).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBorder).
				Padding(0, 1)
)

// NewInteractiveModel creates a ready-to-run bubbletea model.
func NewInteractiveModel(modules []Module) interactiveModel {
	ti := textinput.New()
	ti.Placeholder = "type to filter..."
	ti.CharLimit = 64

	return interactiveModel{
		modules:  modules,
		filtered: modules,
		search:   ti,
		mode:     modeOverview,
	}
}

func (m interactiveModel) Init() tea.Cmd {
	return nil
}

// applyFilter returns modules (and within them, commands) matching the query.
func applyFilter(modules []Module, query string) []Module {
	if query == "" {
		return modules
	}
	lower := strings.ToLower(query)
	var result []Module
	for _, mod := range modules {
		// Check if group name matches
		if strings.Contains(strings.ToLower(mod.Name), lower) ||
			strings.Contains(strings.ToLower(mod.Description), lower) {
			result = append(result, mod)
			continue
		}
		// Check if any command matches
		var matched []Command
		for _, cmd := range mod.Commands {
			if strings.Contains(strings.ToLower(cmd.Name), lower) ||
				strings.Contains(strings.ToLower(cmd.Description), lower) {
				matched = append(matched, cmd)
			}
		}
		if len(matched) > 0 {
			result = append(result, Module{
				Name:        mod.Name,
				Description: mod.Description,
				Commands:    matched,
			})
		}
	}
	return result
}

func (m interactiveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch m.mode {

		case modeOverview:
			return m.updateOverview(msg)

		case modeDetail:
			return m.updateDetail(msg)

		case modeSearch:
			return m.updateSearch(msg)
		}
	}
	return m, nil
}

func (m interactiveModel) updateOverview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}

	case "enter", " ":
		if len(m.filtered) > 0 {
			m.detail = m.filtered[m.cursor]
			m.mode = modeDetail
		}

	case "/":
		m.mode = modeSearch
		m.search.Focus()
		return m, textinput.Blink

	case "esc":
		// Clear search and reset
		m.search.SetValue("")
		m.filtered = m.modules
		m.cursor = 0
	}
	return m, nil
}

func (m interactiveModel) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "backspace", "h":
		m.mode = modeOverview
	}
	return m, nil
}

func (m interactiveModel) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.mode = modeOverview
		m.search.Blur()
		// Keep the filter active but leave search mode

	case "enter":
		if len(m.filtered) > 0 {
			m.detail = m.filtered[m.cursor]
			m.search.Blur()
			m.mode = modeDetail
		}

	case "up", "ctrl+p":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "ctrl+n":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}

	default:
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		query := m.search.Value()
		m.filtered = applyFilter(m.modules, query)
		// Clamp cursor
		if m.cursor >= len(m.filtered) {
			m.cursor = max(0, len(m.filtered)-1)
		}
		return m, cmd
	}
	return m, nil
}

func (m interactiveModel) View() string {
	switch m.mode {
	case modeDetail:
		return m.viewDetail()
	default:
		return m.viewOverview()
	}
}

func (m interactiveModel) viewOverview() string {
	var sb strings.Builder

	// Banner (compact single-line version for interactive mode)
	bannerStyle := lipgloss.NewStyle().Foreground(colorGroupName).Bold(true)
	sb.WriteString(bannerStyle.Render("D.O.T.S"))
	sb.WriteString("  ")
	sb.WriteString(styleDim.Render("Don't Overthink This Shit"))
	sb.WriteString("\n\n")

	// Search bar
	query := m.search.Value()
	if m.mode == modeSearch {
		sb.WriteString(styleSearchPrompt.Render("/") + " " + m.search.View())
	} else if query != "" {
		sb.WriteString(styleDim.Render("filter: ") + styleHighlight.Render(query))
	}
	if m.mode == modeSearch || query != "" {
		sb.WriteString("\n\n")
	}

	// Group list
	if len(m.filtered) == 0 {
		sb.WriteString(styleDim.Render("  no groups match your search"))
		sb.WriteString("\n")
	} else {
		// Find max name length for alignment
		maxLen := 0
		for _, mod := range m.filtered {
			if len(mod.Name) > maxLen {
				maxLen = len(mod.Name)
			}
		}

		for i, mod := range m.filtered {
			count := len(mod.Commands)
			plural := "commands"
			if count == 1 {
				plural = "command"
			}

			namePadded := fmt.Sprintf("%-*s", maxLen, mod.Name)
			countStr := fmt.Sprintf("%2d %s", count, plural)

			var line string
			if i == m.cursor {
				// Selected row
				arrow := styleHighlight.Render(">")
				name := lipgloss.NewStyle().Foreground(colorGroupName).Bold(true).Render(namePadded)
				cnt := lipgloss.NewStyle().Foreground(colorCount).Bold(true).Render(countStr)
				line = fmt.Sprintf("%s %s  %s", arrow, name, cnt)
			} else {
				name := styleGroupHeader.Render(namePadded)
				cnt := styleCount.Render(countStr)
				line = fmt.Sprintf("  %s  %s", name, cnt)
			}

			// Highlight search term in group name if applicable
			if query != "" && m.mode != modeSearch {
				// Already handled by filtered list; just render normally
			}

			sb.WriteString(line + "\n")
		}
	}

	// Status bar
	sb.WriteString("\n")
	var hints string
	if m.mode == modeSearch {
		hints = "↑/↓ navigate  enter drill-down  esc exit search  ctrl+c quit"
	} else {
		hints = "↑/↓ or j/k navigate  enter drill-down  / search  esc clear  q quit"
	}
	sb.WriteString(styleStatusBar.Render(hints))

	return sb.String()
}

func (m interactiveModel) viewDetail() string {
	var sb strings.Builder

	// Header
	mod := m.detail
	headerText := styleGroupHeader.Render(mod.Name)
	if mod.Description != "" {
		headerText += "  " + styleDim.Render("— "+mod.Description)
	}
	sb.WriteString(styleDetailHeader.Render(headerText))
	sb.WriteString("\n\n")

	// Commands
	if len(mod.Commands) == 0 {
		sb.WriteString(styleDim.Render("  (no commands documented)"))
		sb.WriteString("\n")
	} else {
		maxLen := 0
		for _, cmd := range mod.Commands {
			if len(cmd.Name) > maxLen {
				maxLen = len(cmd.Name)
			}
		}
		for _, cmd := range mod.Commands {
			padding := maxLen - len(cmd.Name)
			nameRendered := renderCmdName(cmd.Name) + strings.Repeat(" ", padding)
			var descRendered string
			if cmd.Description != "" {
				descRendered = styleDesc.Render("  " + cmd.Description)
			}
			sb.WriteString(fmt.Sprintf("  %s%s\n", nameRendered, descRendered))
		}
	}

	// Status bar
	sb.WriteString("\n")
	sb.WriteString(styleStatusBar.Render("esc/backspace go back  q quit"))

	return sb.String()
}

// max returns the larger of two ints (Go 1.21 doesn't have built-in max for int).
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RunInteractive launches the bubbletea interactive UI and blocks until the user quits.
func RunInteractive(modules []Module) error {
	model := NewInteractiveModel(modules)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
