package dots

import (
	"sort"
	"strings"
)

// Auto-commit subjects say what changed and where, not when — git already
// records the timestamp. A log line like "🍺 personal: brew packages" is
// scannable across machines in a way "sync: 2026-07-19" never was.

// changeCategory is one bucket of related paths, with the emoji and label
// used when a commit touches only that bucket.
type changeCategory struct {
	emoji string
	label string
	// match reports whether a repo-relative path belongs to this category.
	match func(path string) bool
}

// mixedEmoji marks a commit that spans more than one category.
const mixedEmoji = "🎲"

// fallbackCategory covers paths no other category claims.
var fallbackCategory = changeCategory{emoji: "🟣", label: "odds and ends"}

// categories are tried in order; the first match wins, so narrower rules
// (Brewfile, p10k.zsh) must precede broader ones (machines/, *.zsh).
var categories = []changeCategory{
	{"🍺", "brew packages", func(p string) bool {
		return base(p) == "Brewfile"
	}},
	{"📦", "package lists", func(p string) bool {
		b := base(p)
		return b == "npm_list" || b == "python_packages"
	}},
	{"🐙", "git config", func(p string) bool {
		return base(p) == "gitconfig"
	}},
	{"💅", "prompt", func(p string) bool {
		return base(p) == "p10k.zsh"
	}},
	{"🔧", "tools", func(p string) bool {
		return hasDir(p, "tools") || base(p) == "Makefile"
	}},
	{"🪝", "git hooks", func(p string) bool {
		return hasDir(p, "hooks")
	}},
	{"🧿", "issues", func(p string) bool {
		return hasDir(p, ".beads")
	}},
	{"📖", "docs", func(p string) bool {
		return strings.HasSuffix(p, ".md")
	}},
	{"🐚", "shell config", func(p string) bool {
		return strings.HasSuffix(p, ".zsh") || base(p) == "zshrc" ||
			hasDir(p, "shell") || hasDir(p, "modules")
	}},
	{"⚙️", "machine config", func(p string) bool {
		return hasDir(p, "machines")
	}},
}

// commitSubject builds the auto-commit subject from the paths a sync touched.
// One category commits get that category's emoji and label; multi-category
// commits get the mixed emoji and up to maxLabels labels.
func commitSubject(machine string, paths []string) string {
	if machine == "" {
		machine = "unknown"
	}

	cats := classify(paths)
	if len(cats) == 0 {
		return fallbackCategory.emoji + " " + machine + ": changes"
	}

	emoji := mixedEmoji
	if len(cats) == 1 {
		emoji = cats[0].emoji
	}
	return emoji + " " + machine + ": " + summarize(cats)
}

// classify maps paths to their categories, preserving the declaration order
// of categories so subjects read the same way for the same kind of change.
func classify(paths []string) []changeCategory {
	seen := map[string]bool{}
	var out []changeCategory

	add := func(c changeCategory) {
		if seen[c.label] {
			return
		}
		seen[c.label] = true
		out = append(out, c)
	}

	for _, p := range paths {
		matched := false
		for _, c := range categories {
			if c.match(p) {
				add(c)
				matched = true
				break
			}
		}
		if !matched {
			add(fallbackCategory)
		}
	}

	sort.SliceStable(out, func(i, j int) bool {
		return categoryRank(out[i]) < categoryRank(out[j])
	})
	return out
}

// categoryRank orders categories by their position in categories, with the
// fallback last so "odds and ends" never leads a subject.
func categoryRank(c changeCategory) int {
	for i, known := range categories {
		if known.label == c.label {
			return i
		}
	}
	return len(categories)
}

// maxLabels caps how many labels a subject lists before eliding the rest,
// keeping subjects under a readable width in git log --oneline.
const maxLabels = 3

// summarize joins category labels, eliding past maxLabels.
func summarize(cats []changeCategory) string {
	labels := make([]string, 0, len(cats))
	for _, c := range cats {
		labels = append(labels, c.label)
	}
	if len(labels) <= maxLabels {
		return strings.Join(labels, ", ")
	}
	rest := len(labels) - maxLabels
	suffix := " + 1 more"
	if rest > 1 {
		suffix = " + " + itoa(rest) + " more"
	}
	return strings.Join(labels[:maxLabels], ", ") + suffix
}

// parseStatusPaths extracts repo-relative paths from git status --porcelain
// output. Rename entries ("R  old -> new") yield the destination path.
func parseStatusPaths(porcelain string) []string {
	var paths []string
	for _, line := range strings.Split(porcelain, "\n") {
		if len(line) < 4 {
			continue
		}
		// Porcelain v1: two status columns, a space, then the path.
		p := strings.TrimSpace(line[3:])
		if i := strings.Index(p, " -> "); i >= 0 {
			p = p[i+4:]
		}
		p = strings.Trim(p, `"`)
		if p != "" {
			paths = append(paths, p)
		}
	}
	return paths
}

// base returns the final path element.
func base(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

// hasDir reports whether dir appears as a path component of p.
func hasDir(p, dir string) bool {
	for _, part := range strings.Split(p, "/") {
		if part == dir {
			return true
		}
	}
	return false
}

// itoa converts small non-negative ints without pulling in strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
