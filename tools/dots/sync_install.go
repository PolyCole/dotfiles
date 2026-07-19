package dots

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const plistLabel = "com.dotfiles.sync"
const plistFilename = "com.dotfiles.sync.plist"

var plistTmpl = template.Must(template.New("plist").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>{{.Label}}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{.DotsBin}}</string>
		<string>sync</string>
		<string>now</string>
	</array>
	<key>EnvironmentVariables</key>
	<dict>
		<key>PATH</key>
		<string>{{.Path}}</string>
		<key>DOTFILES</key>
		<string>{{.Dotfiles}}</string>
		<key>DOTFILES_MACHINE</key>
		<string>{{.Machine}}</string>
		<key>HOME</key>
		<string>{{.Home}}</string>
	</dict>
	<key>StartCalendarInterval</key>
	<dict>
		<key>Hour</key>
		<integer>9</integer>
		<key>Minute</key>
		<integer>0</integer>
	</dict>
	<key>StandardOutPath</key>
	<string>{{.LogPath}}</string>
	<key>StandardErrorPath</key>
	<string>{{.LogPath}}</string>
</dict>
</plist>
`))

type plistData struct {
	Label    string
	DotsBin  string
	Path     string
	Dotfiles string
	Machine  string
	Home     string
	LogPath  string
}

// RunSyncInstall installs the launchd job that runs 'dots sync now' daily.
// It generates the plist, loads it via launchctl, and then runs 'dots sync link'.
func RunSyncInstall(w io.Writer, dotfiles, machine string) error {
	if machine == "" {
		fmt.Fprintln(w, styleSyncErr.Render("error:"), styleDim.Render("DOTFILES_MACHINE is not set"))
		return fmt.Errorf("DOTFILES_MACHINE is not set")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	fmt.Fprintf(w, "\n%s\n\n", styleSyncLabel.Render("dots sync install"))

	// Check for existing crontab entry
	warnCrontab(w)

	// Rebuild binaries first so the launchd job never loads a stale binary.
	buildBinaries(w, dotfiles)

	// Resolve the dots binary (the running executable)
	dotsBin, err := os.Executable()
	if err != nil {
		dotsBin = filepath.Join(dotfiles, "bin", "dots")
	}
	// Resolve any symlinks so the plist points at the real binary
	if resolved, err := filepath.EvalSymlinks(dotsBin); err == nil {
		dotsBin = resolved
	}

	plistDir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(plistDir, 0o755); err != nil {
		return fmt.Errorf("cannot create LaunchAgents directory: %w", err)
	}
	plistPath := filepath.Join(plistDir, plistFilename)

	// Build the PATH from the current environment, falling back to a safe default
	envPath := os.Getenv("PATH")
	if envPath == "" {
		envPath = "/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
	}

	data := plistData{
		Label:    plistLabel,
		DotsBin:  dotsBin,
		Path:     envPath,
		Dotfiles: dotfiles,
		Machine:  machine,
		Home:     home,
		LogPath:  syncLogPath(),
	}

	var buf bytes.Buffer
	if err := plistTmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("cannot render plist template: %w", err)
	}

	if err := os.WriteFile(plistPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("cannot write plist: %w", err)
	}
	fmt.Fprintf(w, "  %s  %s\n", styleSyncOK.Render("write"), styleSyncPath.Render(plistPath))

	// Unload first in case an old version is loaded
	unloadCmd := exec.Command("launchctl", "unload", plistPath)
	_ = unloadCmd.Run() // ignore error — it may not be loaded yet

	loadCmd := exec.Command("launchctl", "load", plistPath)
	var loadOut bytes.Buffer
	loadCmd.Stderr = &loadOut
	if err := loadCmd.Run(); err != nil {
		msg := strings.TrimSpace(loadOut.String())
		fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleDim.Render("launchctl load failed: "+msg))
		return fmt.Errorf("launchctl load failed: %w", err)
	}
	fmt.Fprintf(w, "  %s  %s\n", styleSyncOK.Render("load "), styleDim.Render(plistLabel))

	// Run sync link to set up symlinks
	fmt.Fprintln(w)
	if err := RunSyncLink(w, dotfiles, machine); err != nil {
		return err
	}

	fmt.Fprintf(w, "  %s  %s\n",
		styleSyncOK.Render("done "),
		styleDim.Render("dots sync now will run daily at 09:00 — logs at ~/.dotfiles-sync.log"),
	)
	fmt.Fprintln(w)
	return nil
}

// RunSyncUninstall unloads and removes the launchd plist.
// Symlinks are left in place.
func RunSyncUninstall(w io.Writer) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	fmt.Fprintf(w, "\n%s\n\n", styleSyncLabel.Render("dots sync uninstall"))

	plistPath := filepath.Join(home, "Library", "LaunchAgents", plistFilename)

	// Unload (best-effort — may not be loaded)
	unloadCmd := exec.Command("launchctl", "unload", plistPath)
	var unloadOut bytes.Buffer
	unloadCmd.Stderr = &unloadOut
	if err := unloadCmd.Run(); err != nil {
		// Only warn if the file existed — otherwise it was never installed
		if _, statErr := os.Stat(plistPath); statErr == nil {
			msg := strings.TrimSpace(unloadOut.String())
			fmt.Fprintf(w, "  %s  %s\n", styleSyncWarn.Render("warn "), styleDim.Render("launchctl unload: "+msg))
		}
	} else {
		fmt.Fprintf(w, "  %s  %s\n", styleSyncOK.Render("unload"), styleDim.Render(plistLabel))
	}

	// Remove plist
	if err := os.Remove(plistPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(w, "  %s  %s\n", styleSyncSkip.Render("skip "), styleDim.Render("plist not found — already uninstalled"))
		} else {
			fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleDim.Render("cannot remove plist: "+err.Error()))
			return fmt.Errorf("cannot remove plist: %w", err)
		}
	} else {
		fmt.Fprintf(w, "  %s  %s\n", styleSyncOK.Render("remove"), styleSyncPath.Render(plistPath))
	}

	fmt.Fprintf(w, "\n  %s\n", styleDim.Render("symlinks were not removed — run 'dots sync link' to re-apply"))
	fmt.Fprintln(w)
	return nil
}

// buildBinaries runs 'make all' so a fresh install never loads a stale binary.
// A build failure warns but does not abort the install — matching the
// non-fatal philosophy of rebuildIfChanged in the sync cycle.
func buildBinaries(w io.Writer, dotfiles string) {
	cmd := exec.Command("make", "all")
	cmd.Dir = dotfiles
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(w, "  %s  %s\n", styleSyncWarn.Render("warn "), styleDim.Render("make all failed: "+err.Error()))
		if msg := strings.TrimSpace(out.String()); msg != "" {
			fmt.Fprintf(w, "         %s\n", styleDim.Render(msg))
		}
		return
	}
	fmt.Fprintf(w, "  %s  %s\n", styleSyncOK.Render("build"), styleDim.Render("rebuilt bin/dots and bin/startup-message"))
}

// warnCrontab prints a warning if a crontab entry related to dotfiles sync exists.
func warnCrontab(w io.Writer) {
	cmd := exec.Command("crontab", "-l")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return // no crontab or crontab not available
	}

	for _, line := range strings.Split(out.String(), "\n") {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "dotfiles") || strings.Contains(lower, "dots") || strings.Contains(lower, "sync") {
			fmt.Fprintf(w, "  %s  %s\n",
				styleSyncWarn.Render("warn "),
				styleDim.Render("existing crontab entry may conflict with launchd — consider removing it:"),
			)
			fmt.Fprintf(w, "         %s\n", styleSyncPath.Render(strings.TrimSpace(line)))
			fmt.Fprintf(w, "         %s\n\n", styleDim.Render("run: crontab -e"))
			return
		}
	}
}
