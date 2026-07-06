package dots

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const launchdLabel = "com.dotfiles.sync"

// launchdPlistPath returns the expected path of the launchd plist.
func launchdPlistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", launchdLabel+".plist")
}

// syncLogPath returns the log file the launchd job writes to.
// Shared by install (which sets it in the plist) and status (which reads it).
func syncLogPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".dotfiles-sync.log")
}

// launchdStatus holds information retrieved from launchctl.
type launchdStatus struct {
	installed    bool
	loaded       bool
	lastExitCode string // "-" if never run
	pid          string // "-" if not running
}

// queryLaunchd checks plist existence and queries launchctl list.
func queryLaunchd() launchdStatus {
	plist := launchdPlistPath()
	_, statErr := os.Stat(plist)
	installed := statErr == nil

	out, err := exec.Command("launchctl", "list", launchdLabel).Output()
	if err != nil || len(out) == 0 {
		return launchdStatus{installed: installed, loaded: false, lastExitCode: "-", pid: "-"}
	}

	// Parse the launchctl list output (tab-separated: PID, LastExitStatus, Label)
	// Single-line format: <pid>\t<exit>\t<label>
	line := strings.TrimSpace(string(out))
	parts := strings.Fields(line)
	pid := "-"
	exitCode := "-"
	if len(parts) >= 2 {
		pid = parts[0]
		exitCode = parts[1]
	}
	return launchdStatus{
		installed:    installed,
		loaded:       true,
		lastExitCode: exitCode,
		pid:          pid,
	}
}

// intAfterKey extracts the first <integer> value following <key>name</key> in s.
func intAfterKey(s, name string) (int, bool) {
	idx := strings.Index(s, "<key>"+name+"</key>")
	if idx == -1 {
		return 0, false
	}
	after := s[idx:]
	start := strings.Index(after, "<integer>")
	end := strings.Index(after, "</integer>")
	if start == -1 || end == -1 || end <= start {
		return 0, false
	}
	var n int
	if _, err := fmt.Sscanf(strings.TrimSpace(after[start+len("<integer>"):end]), "%d", &n); err != nil {
		return 0, false
	}
	return n, true
}

// readPlistInterval attempts to extract the StartInterval (seconds) from the plist.
// Returns 0 if it cannot be parsed.
func readPlistInterval() int {
	data, err := os.ReadFile(launchdPlistPath())
	if err != nil {
		return 0
	}
	n, _ := intAfterKey(string(data), "StartInterval")
	return n
}

// readPlistCalendar extracts the StartCalendarInterval Hour/Minute from the
// plist — the schedule format 'dots sync install' actually writes.
func readPlistCalendar() (hour, minute int, ok bool) {
	data, err := os.ReadFile(launchdPlistPath())
	if err != nil {
		return 0, 0, false
	}
	content := string(data)
	idx := strings.Index(content, "<key>StartCalendarInterval</key>")
	if idx == -1 {
		return 0, 0, false
	}
	section := content[idx:]
	if end := strings.Index(section, "</dict>"); end != -1 {
		section = section[:end]
	}
	hour, hourOK := intAfterKey(section, "Hour")
	minute, minOK := intAfterKey(section, "Minute")
	if !hourOK && !minOK {
		return 0, 0, false
	}
	return hour, minute, true
}

// readLastRunFromLog attempts to read the last run timestamp from the sync
// log file. Each 'dots sync now' run logs a line starting with an RFC3339
// timestamp.
func readLastRunFromLog() (time.Time, bool) {
	data, err := os.ReadFile(syncLogPath())
	if err != nil {
		return time.Time{}, false
	}
	// Scan lines in reverse to find last timestamp
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		// Try parsing first 20+ chars as a timestamp
		for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05"} {
			prefix := line
			if len(prefix) > len(layout) {
				prefix = prefix[:len(layout)]
			}
			t, err := time.Parse(layout, prefix)
			if err == nil {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

// symlinkStatus describes the state of a managed config's symlink.
type symlinkStatus struct {
	source string
	target string
	state  string // "correct", "missing", "broken", "conflict"
}

// checkSymlinks inspects each config entry in the manifest.
func checkSymlinks(dotfiles string, manifest *SyncManifest) []symlinkStatus {
	results := make([]symlinkStatus, 0, len(manifest.Configs))
	for _, cfg := range manifest.Configs {
		srcAbs := filepath.Join(dotfiles, cfg.Source)
		tgtAbs, err := expandHome(cfg.Target)
		if err != nil {
			results = append(results, symlinkStatus{cfg.Source, cfg.Target, "conflict"})
			continue
		}

		info, lstatErr := os.Lstat(tgtAbs)
		if os.IsNotExist(lstatErr) {
			results = append(results, symlinkStatus{cfg.Source, cfg.Target, "missing"})
			continue
		}
		if lstatErr != nil {
			results = append(results, symlinkStatus{cfg.Source, cfg.Target, "conflict"})
			continue
		}

		if info.Mode()&os.ModeSymlink != 0 {
			dest, readErr := os.Readlink(tgtAbs)
			if readErr != nil {
				results = append(results, symlinkStatus{cfg.Source, cfg.Target, "broken"})
				continue
			}
			// Check if the symlink target exists
			if _, statErr := os.Stat(tgtAbs); statErr != nil {
				results = append(results, symlinkStatus{cfg.Source, cfg.Target, "broken"})
				continue
			}
			if dest == srcAbs {
				results = append(results, symlinkStatus{cfg.Source, cfg.Target, "correct"})
			} else {
				results = append(results, symlinkStatus{cfg.Source, cfg.Target, "conflict"})
			}
		} else {
			// Regular file exists at target — conflict
			results = append(results, symlinkStatus{cfg.Source, cfg.Target, "conflict"})
		}
	}
	return results
}

// snapshotInfo holds the last-captured timestamp for a snapshot file.
type snapshotInfo struct {
	dest    string
	command string
	lastRun time.Time
	hasRun  bool
}

// querySnapshotTimes uses git log to find the last commit time for each snapshot file.
func querySnapshotTimes(dotfiles string, manifest *SyncManifest) []snapshotInfo {
	results := make([]snapshotInfo, 0, len(manifest.Snapshots))
	for _, snap := range manifest.Snapshots {
		relPath := snap.Dest
		absPath := filepath.Join(dotfiles, relPath)

		// Check file exists at all
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			results = append(results, snapshotInfo{snap.Dest, snap.Command, time.Time{}, false})
			continue
		}

		// git log --format=%ci -1 -- <file>
		cmd := exec.Command("git", "-C", dotfiles, "log", "--format=%ci", "-1", "--", relPath)
		var buf bytes.Buffer
		cmd.Stdout = &buf
		if err := cmd.Run(); err != nil {
			results = append(results, snapshotInfo{snap.Dest, snap.Command, time.Time{}, false})
			continue
		}
		ts := strings.TrimSpace(buf.String())
		if ts == "" {
			// File exists but no commits — untracked
			results = append(results, snapshotInfo{snap.Dest, snap.Command, time.Time{}, false})
			continue
		}
		t, err := time.Parse("2006-01-02 15:04:05 -0700", ts)
		if err != nil {
			results = append(results, snapshotInfo{snap.Dest, snap.Command, time.Time{}, false})
			continue
		}
		results = append(results, snapshotInfo{snap.Dest, snap.Command, t, true})
	}
	return results
}

// formatAge returns a human-readable age string for a time in the past.
func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case d < 24*time.Hour:
		hrs := int(d.Hours())
		if hrs == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hrs)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// RunSyncStatus displays the status of the dots sync subsystem.
func RunSyncStatus(w io.Writer, dotfiles, machine string) error {
	fmt.Fprintf(w, "\n%s\n\n", styleSyncLabel.Render("dots sync status"))

	// ── Section 1: launchd job ────────────────────────────────────────────
	fmt.Fprintf(w, "%s\n", styleGroupHeader.Render("launchd job"))

	status := queryLaunchd()
	plistPath := launchdPlistPath()

	if !status.installed {
		fmt.Fprintf(w, "  %s  %s\n", styleSyncWarn.Render("not installed"), styleDim.Render(plistPath))
	} else if !status.loaded {
		fmt.Fprintf(w, "  %s      %s\n", styleSyncWarn.Render("installed"), styleSyncPath.Render(plistPath))
		fmt.Fprintf(w, "  %s     %s\n", styleSyncErr.Render("not loaded"), styleDim.Render("run: launchctl load "+plistPath))
	} else {
		fmt.Fprintf(w, "  %s         %s\n", styleSyncOK.Render("loaded"), styleSyncPath.Render(plistPath))
	}

	// ── Section 2: last run ───────────────────────────────────────────────
	fmt.Fprintf(w, "\n%s\n", styleGroupHeader.Render("last run"))

	lastRun, hasLastRun := readLastRunFromLog()
	if hasLastRun {
		fmt.Fprintf(w, "  %s  %s\n",
			styleSyncOK.Render(lastRun.Format("2006-01-02 15:04:05")),
			styleDim.Render("("+formatAge(lastRun)+")"),
		)
	} else if status.loaded && status.lastExitCode != "-" {
		fmt.Fprintf(w, "  %s  %s %s\n",
			styleDim.Render("unknown"),
			styleDim.Render("last exit code:"),
			styleSyncPath.Render(status.lastExitCode),
		)
	} else {
		fmt.Fprintf(w, "  %s\n", styleDim.Render("no runs recorded yet"))
	}

	// ── Section 3: next scheduled run ────────────────────────────────────
	fmt.Fprintf(w, "\n%s\n", styleGroupHeader.Render("next run"))

	interval := readPlistInterval()
	if !status.loaded {
		fmt.Fprintf(w, "  %s\n", styleDim.Render("n/a — job not loaded"))
	} else if hour, minute, ok := readPlistCalendar(); ok {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}
		fmt.Fprintf(w, "  %s  %s\n",
			styleSyncPath.Render(next.Format("2006-01-02 15:04:05")),
			styleDim.Render(fmt.Sprintf("(daily at %02d:%02d)", hour, minute)),
		)
	} else if interval > 0 && hasLastRun {
		next := lastRun.Add(time.Duration(interval) * time.Second)
		if time.Now().After(next) {
			fmt.Fprintf(w, "  %s\n", styleSyncPath.Render("pending (overdue)"))
		} else {
			fmt.Fprintf(w, "  %s  %s\n",
				styleSyncPath.Render(next.Format("2006-01-02 15:04:05")),
				styleDim.Render("(in "+formatAge(time.Now().Add(time.Until(next)))+")"),
			)
		}
	} else if interval > 0 {
		hrs := interval / 3600
		mins := (interval % 3600) / 60
		var intervalStr string
		switch {
		case hrs > 0 && mins > 0:
			intervalStr = fmt.Sprintf("every %dh%dm", hrs, mins)
		case hrs > 0:
			intervalStr = fmt.Sprintf("every %dh", hrs)
		default:
			intervalStr = fmt.Sprintf("every %dm", mins)
		}
		fmt.Fprintf(w, "  %s  %s\n", styleSyncPath.Render(intervalStr), styleDim.Render("(no prior run to compute next)"))
	} else {
		fmt.Fprintf(w, "  %s\n", styleDim.Render("schedule unknown (no schedule found in plist)"))
	}

	// ── Sections 4 & 5 require a manifest ────────────────────────────────
	if machine == "" {
		fmt.Fprintf(w, "\n%s\n", styleDim.Render("DOTFILES_MACHINE not set — skipping config and snapshot status"))
		return nil
	}

	manifestPath := filepath.Join(dotfiles, "machines", machine, "sync.yml")
	manifest, err := LoadSyncManifest(manifestPath)
	if err != nil {
		fmt.Fprintf(w, "\n%s  %s\n", styleSyncWarn.Render("warning:"), styleDim.Render(err.Error()))
		return nil
	}

	// ── Section 4: managed configs ────────────────────────────────────────
	fmt.Fprintf(w, "\n%s\n", styleGroupHeader.Render("managed configs"))

	symlinks := checkSymlinks(dotfiles, manifest)
	if len(symlinks) == 0 {
		fmt.Fprintf(w, "  %s\n", styleDim.Render("none configured"))
	} else {
		maxLen := 0
		for _, sl := range symlinks {
			if len(sl.target) > maxLen {
				maxLen = len(sl.target)
			}
		}
		for _, sl := range symlinks {
			var stateStr string
			switch sl.state {
			case "correct":
				stateStr = styleSyncOK.Render("correct ")
			case "missing":
				stateStr = styleSyncErr.Render("missing ")
			case "broken":
				stateStr = styleSyncErr.Render("broken  ")
			case "conflict":
				stateStr = styleSyncWarn.Render("conflict")
			default:
				stateStr = styleDim.Render("unknown ")
			}
			padding := maxLen - len(sl.target)
			fmt.Fprintf(w, "  %s  %s%s  %s  %s\n",
				stateStr,
				styleSyncPath.Render(sl.target),
				strings.Repeat(" ", padding),
				styleDim.Render("→"),
				styleDim.Render(sl.source),
			)
		}
	}

	// ── Section 5: snapshots ──────────────────────────────────────────────
	fmt.Fprintf(w, "\n%s\n", styleGroupHeader.Render("snapshots"))

	snapshots := querySnapshotTimes(dotfiles, manifest)
	if len(snapshots) == 0 {
		fmt.Fprintf(w, "  %s\n", styleDim.Render("none configured"))
	} else {
		maxLen := 0
		for _, sn := range snapshots {
			if len(sn.dest) > maxLen {
				maxLen = len(sn.dest)
			}
		}
		for _, sn := range snapshots {
			padding := maxLen - len(sn.dest)
			if sn.hasRun {
				fmt.Fprintf(w, "  %s%s  %s  %s\n",
					styleSyncPath.Render(sn.dest),
					strings.Repeat(" ", padding),
					styleSyncOK.Render(sn.lastRun.Format("2006-01-02 15:04:05")),
					styleDim.Render("("+formatAge(sn.lastRun)+")"),
				)
			} else {
				fmt.Fprintf(w, "  %s%s  %s\n",
					styleSyncPath.Render(sn.dest),
					strings.Repeat(" ", padding),
					styleDim.Render("never captured"),
				)
			}
		}
	}

	fmt.Fprintln(w)
	return nil
}
