package dots

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// sync output colors
var (
	colorSyncOK   = lipgloss.Color("#9ECE6A") // green — created
	colorSyncSkip = lipgloss.Color("#565F89") // dim — already linked
	colorSyncWarn = lipgloss.Color("#E0AF68") // amber — conflict/backup
	colorSyncErr  = lipgloss.Color("#F7768E") // red — error
	colorSyncPath = lipgloss.Color("#A9B1D6") // muted blue-gray — path text
)

var (
	styleSyncOK    = lipgloss.NewStyle().Foreground(colorSyncOK).Bold(true)
	styleSyncSkip  = lipgloss.NewStyle().Foreground(colorSyncSkip)
	styleSyncWarn  = lipgloss.NewStyle().Foreground(colorSyncWarn).Bold(true)
	styleSyncErr   = lipgloss.NewStyle().Foreground(colorSyncErr).Bold(true)
	styleSyncLabel = lipgloss.NewStyle().Foreground(colorGroupName).Bold(true)
	styleSyncPath  = lipgloss.NewStyle().Foreground(colorSyncPath)
)

// SyncConfig represents a single config entry in sync.yml.
type SyncConfig struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

// SyncManifest represents the full sync.yml file.
type SyncManifest struct {
	Configs []SyncConfig `yaml:"configs"`
}

// LoadSyncManifest reads and parses a sync.yml file.
func LoadSyncManifest(path string) (*SyncManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m SyncManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse error in %s: %w", path, err)
	}
	return &m, nil
}

// expandHome replaces a leading "~/" with the user's home directory.
func expandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, path[2:]), nil
}

type linkResult int

const (
	linkCreated linkResult = iota
	linkSkipped
	linkBacked
	linkErrored
)

// RunSyncLink performs the 'dots sync link' operation.
// dotfiles is the path to the dotfiles repository root.
// machine is the active machine profile.
func RunSyncLink(w io.Writer, dotfiles, machine string) error {
	if machine == "" {
		fmt.Fprintln(w, styleSyncErr.Render("error:"), styleDim.Render("DOTFILES_MACHINE is not set — cannot determine which sync.yml to use"))
		return fmt.Errorf("DOTFILES_MACHINE is not set")
	}

	manifestPath := filepath.Join(dotfiles, "machines", machine, "sync.yml")
	manifest, err := LoadSyncManifest(manifestPath)
	if err != nil {
		fmt.Fprintln(w, styleSyncErr.Render("error:"), styleDim.Render(err.Error()))
		return err
	}

	fmt.Fprintf(w, "\n%s  %s\n\n",
		styleSyncLabel.Render("dots sync link"),
		styleDim.Render("— "+machine),
	)

	created, skipped, backed, errs := runLinkAll(w, dotfiles, manifest)

	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %s %s   %s %s   %s %s\n\n",
		styleSyncOK.Render(fmt.Sprintf("%d", created)),
		styleDim.Render("created"),
		styleSyncSkip.Render(fmt.Sprintf("%d", skipped)),
		styleDim.Render("skipped"),
		styleSyncWarn.Render(fmt.Sprintf("%d", backed)),
		styleDim.Render("backed up"),
	)

	if len(errs) > 0 {
		return fmt.Errorf("%d error(s) during sync link", len(errs))
	}
	return nil
}

// runLinkAll iterates over all configs and creates symlinks, returning counts.
func runLinkAll(w io.Writer, dotfiles string, manifest *SyncManifest) (created, skipped, backed int, errs []string) {
	for _, cfg := range manifest.Configs {
		srcAbs := filepath.Join(dotfiles, cfg.Source)
		tgtAbs, err := expandHome(cfg.Target)
		if err != nil {
			msg := fmt.Sprintf("cannot expand target %q: %v", cfg.Target, err)
			fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleSyncPath.Render(msg))
			errs = append(errs, msg)
			continue
		}

		result, err := performLink(w, srcAbs, tgtAbs, cfg.Source, cfg.Target)
		if err != nil {
			errs = append(errs, err.Error())
		}
		switch result {
		case linkCreated:
			created++
		case linkSkipped:
			skipped++
		case linkBacked:
			backed++
		}
	}
	return
}

// performLink creates a symlink from tgt -> src. Returns the outcome and any error.
func performLink(w io.Writer, srcAbs, tgtAbs, srcDisplay, tgtDisplay string) (linkResult, error) {
	// Verify the source exists
	if _, err := os.Stat(srcAbs); os.IsNotExist(err) {
		msg := fmt.Sprintf("source does not exist: %s", srcAbs)
		fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("miss "), styleSyncPath.Render(msg))
		return linkErrored, fmt.Errorf("%s", msg)
	}

	info, lstatErr := os.Lstat(tgtAbs)
	if lstatErr == nil {
		// Target exists — determine what to do
		if info.Mode()&os.ModeSymlink != 0 {
			existing, err := os.Readlink(tgtAbs)
			if err == nil && existing == srcAbs {
				// Already correct — skip
				fmt.Fprintf(w, "  %s  %s  →  %s\n",
					styleSyncSkip.Render("skip "),
					styleSyncPath.Render(tgtDisplay),
					styleSyncPath.Render(srcDisplay),
				)
				return linkSkipped, nil
			}
			// Wrong symlink — remove it and fall through to re-create
			if err := os.Remove(tgtAbs); err != nil {
				msg := fmt.Sprintf("cannot remove existing symlink %s: %v", tgtAbs, err)
				fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleSyncPath.Render(msg))
				return linkErrored, fmt.Errorf("%s", msg)
			}
		} else {
			// Regular file or directory — back it up first
			backupPath := tgtAbs + ".bak"
			if err := os.Rename(tgtAbs, backupPath); err != nil {
				msg := fmt.Sprintf("cannot back up %s to %s: %v", tgtAbs, backupPath, err)
				fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleSyncPath.Render(msg))
				return linkErrored, fmt.Errorf("%s", msg)
			}
			fmt.Fprintf(w, "  %s  %s  →  %s\n",
				styleSyncWarn.Render("back "),
				styleSyncPath.Render(tgtDisplay),
				styleSyncPath.Render(tgtDisplay+".bak"),
			)
			if err := createSymlink(w, srcAbs, tgtAbs, srcDisplay, tgtDisplay); err != nil {
				return linkErrored, err
			}
			return linkBacked, nil
		}
	} else if !os.IsNotExist(lstatErr) {
		msg := fmt.Sprintf("cannot stat %s: %v", tgtAbs, lstatErr)
		fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleSyncPath.Render(msg))
		return linkErrored, fmt.Errorf("%s", msg)
	}

	// Target doesn't exist (or stale symlink was removed) — create it
	if err := createSymlink(w, srcAbs, tgtAbs, srcDisplay, tgtDisplay); err != nil {
		return linkErrored, err
	}
	return linkCreated, nil
}

// createSymlink creates the parent directory if needed and then the symlink.
func createSymlink(w io.Writer, srcAbs, tgtAbs, srcDisplay, tgtDisplay string) error {
	if err := os.MkdirAll(filepath.Dir(tgtAbs), 0o755); err != nil {
		msg := fmt.Sprintf("cannot create parent directory for %s: %v", tgtAbs, err)
		fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleSyncPath.Render(msg))
		return fmt.Errorf("%s", msg)
	}

	if err := os.Symlink(srcAbs, tgtAbs); err != nil {
		msg := fmt.Sprintf("cannot create symlink %s: %v", tgtAbs, err)
		fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleSyncPath.Render(msg))
		return fmt.Errorf("%s", msg)
	}

	fmt.Fprintf(w, "  %s  %s  →  %s\n",
		styleSyncOK.Render("link "),
		styleSyncPath.Render(tgtDisplay),
		styleSyncPath.Render(srcDisplay),
	)
	return nil
}
