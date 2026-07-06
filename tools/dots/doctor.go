package dots

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// doctorIssue is a single finding from 'dots doctor'.
type doctorIssue struct {
	severity string // "error" or "warn"
	location string // repo-relative file the issue was found in
	message  string
}

var (
	reAlias    = regexp.MustCompile(`^\s*alias\s+(?:-g\s+)?([^=\s]+)=`)
	reFuncKw   = regexp.MustCompile(`^\s*function\s+([A-Za-z0-9_:.\-]+)`)
	reFuncPosn = regexp.MustCompile(`^\s*([A-Za-z0-9_:.\-]+)\s*\(\)\s*\{?`)
)

// extractDefinedNames returns the names of all aliases and functions defined
// in a module file.
func extractDefinedNames(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	names := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if m := reAlias.FindStringSubmatch(line); m != nil {
			names[m[1]] = true
			continue
		}
		if m := reFuncKw.FindStringSubmatch(line); m != nil {
			names[m[1]] = true
			continue
		}
		if m := reFuncPosn.FindStringSubmatch(line); m != nil {
			names[m[1]] = true
		}
	}
	return names, scanner.Err()
}

// docName reduces a documented command signature to its bare name,
// e.g. "build_and_promote <service> <git-hash>" → "build_and_promote".
func docName(signature string) string {
	fields := strings.Fields(signature)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// checkModule compares a module's # Commands: block against what the file
// actually defines.
func checkModule(dotfiles, path string) []doctorIssue {
	var issues []doctorIssue
	rel, err := filepath.Rel(dotfiles, path)
	if err != nil {
		rel = path
	}

	mod, err := parseFile(path)
	if err != nil {
		return []doctorIssue{{"error", rel, "cannot parse: " + err.Error()}}
	}
	defined, err := extractDefinedNames(path)
	if err != nil {
		return []doctorIssue{{"error", rel, "cannot read: " + err.Error()}}
	}

	if len(mod.Commands) == 0 {
		issues = append(issues, doctorIssue{"warn", rel, "no # Commands: block — module is invisible to dots"})
	}

	documented := make(map[string]bool)
	for _, cmd := range mod.Commands {
		name := docName(cmd.Name)
		if name == "" {
			continue
		}
		documented[name] = true
		if !defined[name] {
			issues = append(issues, doctorIssue{"error", rel,
				fmt.Sprintf("documented command %q is not defined in this file", name)})
		}
	}
	for name := range defined {
		if !documented[name] {
			issues = append(issues, doctorIssue{"warn", rel,
				fmt.Sprintf("%q is defined but missing from the # Commands: block", name)})
		}
	}
	return issues
}

// checkSyncManifests verifies that every config source in each machine's
// sync.yml exists in the repo.
func checkSyncManifests(dotfiles string) []doctorIssue {
	var issues []doctorIssue
	manifests, _ := filepath.Glob(filepath.Join(dotfiles, "machines", "*", "sync.yml"))
	for _, path := range manifests {
		rel, _ := filepath.Rel(dotfiles, path)
		manifest, err := LoadSyncManifest(path)
		if err != nil {
			issues = append(issues, doctorIssue{"error", rel, err.Error()})
			continue
		}
		for _, cfg := range manifest.Configs {
			if _, err := os.Stat(filepath.Join(dotfiles, cfg.Source)); os.IsNotExist(err) {
				issues = append(issues, doctorIssue{"error", rel,
					fmt.Sprintf("config source %q does not exist", cfg.Source)})
			}
		}
	}
	return issues
}

// checkHooksConfs verifies that every script referenced in each machine's
// hooks.conf exists under hooks/<hook-type>/.
func checkHooksConfs(dotfiles string) []doctorIssue {
	var issues []doctorIssue
	confs, _ := filepath.Glob(filepath.Join(dotfiles, "machines", "*", "hooks.conf"))
	for _, path := range confs {
		rel, _ := filepath.Rel(dotfiles, path)
		data, err := os.ReadFile(path)
		if err != nil {
			issues = append(issues, doctorIssue{"error", rel, "cannot read: " + err.Error()})
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) != 2 {
				issues = append(issues, doctorIssue{"error", rel,
					fmt.Sprintf("malformed line %q — expected '<hook-type> <script-name>'", line)})
				continue
			}
			script := filepath.Join(dotfiles, "hooks", fields[0], fields[1]+".sh")
			if _, err := os.Stat(script); os.IsNotExist(err) {
				issues = append(issues, doctorIssue{"error", rel,
					fmt.Sprintf("hook script hooks/%s/%s.sh does not exist", fields[0], fields[1])})
			}
		}
	}
	return issues
}

// RunDoctor lints the whole repo: every module (all machines, not just the
// active one), every sync.yml, and every hooks.conf. Errors mean something
// is broken or misleading; warns are drift worth knowing about.
// Returns a non-nil error when any error-severity issue is found, so it can
// gate CI.
func RunDoctor(w io.Writer, dotfiles string) error {
	fmt.Fprintf(w, "\n%s\n\n", styleSyncLabel.Render("dots doctor"))

	var issues []doctorIssue

	modulePaths, _ := filepath.Glob(filepath.Join(dotfiles, "modules", "*.zsh"))
	machineModules, _ := filepath.Glob(filepath.Join(dotfiles, "machines", "*", "modules", "*.zsh"))
	modulePaths = append(modulePaths, machineModules...)
	for _, path := range modulePaths {
		issues = append(issues, checkModule(dotfiles, path)...)
	}
	issues = append(issues, checkSyncManifests(dotfiles)...)
	issues = append(issues, checkHooksConfs(dotfiles)...)

	errors, warns := 0, 0
	for _, issue := range issues {
		var tag string
		if issue.severity == "error" {
			tag = styleSyncErr.Render("error")
			errors++
		} else {
			tag = styleSyncWarn.Render("warn ")
			warns++
		}
		fmt.Fprintf(w, "  %s  %s  %s\n", tag, styleSyncPath.Render(issue.location), styleDim.Render(issue.message))
	}

	if len(issues) == 0 {
		fmt.Fprintf(w, "  %s\n", styleSyncOK.Render("all checks passed"))
	}
	fmt.Fprintf(w, "\n  %s %s   %s %s\n\n",
		styleSyncErr.Render(fmt.Sprintf("%d", errors)),
		styleDim.Render("errors"),
		styleSyncWarn.Render(fmt.Sprintf("%d", warns)),
		styleDim.Render("warnings"),
	)

	if errors > 0 {
		return fmt.Errorf("%d error(s) found", errors)
	}
	return nil
}
