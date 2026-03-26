package dots

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Command represents a single command entry from a # Commands: block.
type Command struct {
	// Name is the command name (and optional args), e.g. "git-purge-dir <dir>"
	Name string
	// Description is the human-readable description, e.g. "Remove directory from repo history"
	Description string
}

// Module represents a parsed .zsh module file.
type Module struct {
	// Name is the filename without the .zsh extension.
	Name string
	// Description is the header comment (second line of the file if it starts with #).
	Description string
	// Commands are the entries extracted from the # Commands: block.
	Commands []Command
}

// ParseFiles parses each .zsh file at the given paths and returns a slice of
// Module values. Files that cannot be opened are silently skipped.
func ParseFiles(paths []string) []Module {
	modules := make([]Module, 0, len(paths))
	for _, p := range paths {
		m, err := parseFile(p)
		if err != nil {
			continue
		}
		modules = append(modules, m)
	}
	return modules
}

// parseFile parses a single .zsh file.
func parseFile(path string) (Module, error) {
	f, err := os.Open(path)
	if err != nil {
		return Module{}, err
	}
	defer f.Close()

	base := filepath.Base(path)
	name := strings.TrimSuffix(base, ".zsh")

	var description string
	var commands []Command

	scanner := bufio.NewScanner(f)
	lineNum := 0
	inBlock := false

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// The module description is extracted from the second line if it is a
		// comment and the file follows the standard header convention:
		//   # modules/foo.zsh
		//   # Brief description of the module
		if lineNum == 2 && strings.HasPrefix(line, "#") {
			description = strings.TrimSpace(strings.TrimPrefix(line, "#"))
			continue
		}

		// Detect the start of a Commands: block.
		if line == "# Commands:" {
			inBlock = true
			continue
		}

		if inBlock {
			// Block ends at the first line that is not a comment.
			if !strings.HasPrefix(line, "#") {
				inBlock = false
				continue
			}

			// Strip the leading '#' and one optional space.
			entry := strings.TrimPrefix(line, "#")
			if strings.HasPrefix(entry, " ") {
				entry = entry[1:]
			}

			cmd := parseCommandEntry(entry)
			if cmd.Name != "" {
				commands = append(commands, cmd)
			}
		}
	}

	return Module{
		Name:        name,
		Description: description,
		Commands:    commands,
	}, scanner.Err()
}

// parseCommandEntry splits a trimmed command entry line into name and
// description. It handles three separator formats:
//
//   - Arrow:      "cmd-name → Description text"
//   - Dash:       "cmd-name   - Description text"
//   - Space-only: "cmd-name   Description text"  (no separator character)
func parseCommandEntry(entry string) Command {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return Command{}
	}

	// Arrow separator (→, U+2192).
	if idx := strings.Index(entry, "→"); idx != -1 {
		name := strings.TrimSpace(entry[:idx])
		desc := strings.TrimSpace(entry[idx+len("→"):])
		return Command{Name: name, Description: desc}
	}

	// Dash separator: look for " - " (space, dash, space) to avoid matching
	// hyphens inside command names like "git-purge-dir".
	if idx := strings.Index(entry, " - "); idx != -1 {
		name := strings.TrimSpace(entry[:idx])
		desc := strings.TrimSpace(entry[idx+3:])
		return Command{Name: name, Description: desc}
	}

	// Space-only: split on the first run of two or more spaces so that single
	// spaces within a command signature (e.g. "cmd <arg>") are preserved.
	fields := splitOnDoubleSpace(entry)
	if len(fields) >= 2 {
		return Command{Name: strings.TrimSpace(fields[0]), Description: strings.TrimSpace(fields[1])}
	}

	// No separator found — treat the whole entry as the command name.
	return Command{Name: entry}
}

// splitOnDoubleSpace splits s on the first occurrence of two or more
// consecutive spaces, returning at most two parts.
func splitOnDoubleSpace(s string) []string {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == ' ' && s[i+1] == ' ' {
			return []string{s[:i], strings.TrimLeft(s[i:], " ")}
		}
	}
	return []string{s}
}
