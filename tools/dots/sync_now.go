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

// RunSyncNow performs the full sync cycle and sends a desktop notification
// on failure so unattended launchd runs don't fail silently.
func RunSyncNow(w io.Writer, dotfiles, machine string) error {
	err := runSyncNow(w, dotfiles, machine)
	if err != nil {
		notifySyncFailure(err.Error())
	}
	return err
}

// runSyncNow is the sync cycle itself:
//  1. git pull --rebase
//  2. rebuild binaries if the pull brought new commits
//  3. run each snapshot command and write output to dest
//  4. verify config symlinks (warn on broken)
//  5. if changes exist, commit and push; otherwise exit cleanly
func runSyncNow(w io.Writer, dotfiles, machine string) error {
	if machine == "" {
		fmt.Fprintln(w, styleSyncErr.Render("error:"), styleDim.Render("DOTFILES_MACHINE is not set — cannot determine which sync.yml to use"))
		return fmt.Errorf("DOTFILES_MACHINE is not set")
	}

	fmt.Fprintf(w, "\n%s  %s\n",
		styleSyncLabel.Render("dots sync now"),
		styleDim.Render("— "+machine),
	)
	// RFC3339 timestamp so 'dots sync status' can read the last run from the log
	fmt.Fprintf(w, "%s\n\n", styleDim.Render(time.Now().Format(time.RFC3339)))

	// ── Step 1: git pull --rebase ─────────────────────────────────────────
	fmt.Fprintf(w, "%s\n", styleGroupHeader.Render("pull"))
	headBefore := gitHead(dotfiles)
	if err := runGitPullRebase(w, dotfiles); err != nil {
		return err
	}

	// ── Step 2: rebuild binaries if the pull brought new commits ─────────
	fmt.Fprintf(w, "\n%s\n", styleGroupHeader.Render("build"))
	rebuildIfChanged(w, dotfiles, headBefore)

	// ── Step 3: load manifest ─────────────────────────────────────────────
	manifestPath := filepath.Join(dotfiles, "machines", machine, "sync.yml")
	manifest, err := LoadSyncManifest(manifestPath)
	if err != nil {
		fmt.Fprintln(w, styleSyncErr.Render("error:"), styleDim.Render(err.Error()))
		return err
	}

	// ── Step 4: run snapshots ─────────────────────────────────────────────
	fmt.Fprintf(w, "\n%s\n", styleGroupHeader.Render("snapshots"))
	runSnapshots(w, dotfiles, manifest)

	// ── Step 5: verify symlinks ───────────────────────────────────────────
	fmt.Fprintf(w, "\n%s\n", styleGroupHeader.Render("symlinks"))
	checkSymlinksNow(w, dotfiles, manifest)

	// ── Step 6: commit and push if dirty ─────────────────────────────────
	fmt.Fprintf(w, "\n%s\n", styleGroupHeader.Render("commit"))
	if err := commitAndPush(w, dotfiles, machine); err != nil {
		return err
	}

	fmt.Fprintln(w)
	return nil
}

// gitHead returns the current HEAD commit, or "" if it cannot be determined.
func gitHead(dotfiles string) string {
	out, err := exec.Command("git", "-C", dotfiles, "rev-parse", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// rebuildIfChanged runs 'make all' when the pull moved HEAD, so machines
// don't keep running stale binaries after upstream tool changes.
// Build failures warn but do not abort the sync.
func rebuildIfChanged(w io.Writer, dotfiles, headBefore string) {
	if headBefore != "" && gitHead(dotfiles) == headBefore {
		fmt.Fprintf(w, "  %s\n", styleSyncSkip.Render("no new commits — binaries up to date"))
		return
	}

	cmd := exec.Command("make", "all")
	cmd.Dir = dotfiles
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(w, "  %s  %s\n",
			styleSyncWarn.Render("warn "),
			styleDim.Render("make all failed: "+err.Error()),
		)
		if msg := strings.TrimSpace(out.String()); msg != "" {
			fmt.Fprintf(w, "         %s\n", styleDim.Render(msg))
		}
		return
	}
	fmt.Fprintf(w, "  %s\n", styleSyncOK.Render("rebuilt bin/dots and bin/startup-message"))
}

// runGitPullRebase runs git pull --rebase in dotfiles. --autostash lets the
// pull succeed when the working tree is dirty (common on a repo you actively
// edit): local changes are stashed before the rebase and reapplied after, so
// the later commit step still picks them up.
func runGitPullRebase(w io.Writer, dotfiles string) error {
	cmd := exec.Command("git", "-C", dotfiles, "pull", "--rebase", "--autostash")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleDim.Render(out.String()))
		return fmt.Errorf("git pull --rebase failed: %w", err)
	}
	outStr := strings.TrimSpace(out.String())
	if outStr == "Current branch main is up to date." || strings.Contains(outStr, "up to date") {
		fmt.Fprintf(w, "  %s\n", styleSyncSkip.Render("already up to date"))
	} else if outStr != "" {
		fmt.Fprintf(w, "  %s\n", styleDim.Render(outStr))
	} else {
		fmt.Fprintf(w, "  %s\n", styleSyncOK.Render("pulled"))
	}
	return nil
}

// runSnapshots executes each snapshot command and writes output to dest.
// Failed commands warn but do not abort.
func runSnapshots(w io.Writer, dotfiles string, manifest *SyncManifest) {
	if len(manifest.Snapshots) == 0 {
		fmt.Fprintf(w, "  %s\n", styleDim.Render("none configured"))
		return
	}

	for _, snap := range manifest.Snapshots {
		destAbs := filepath.Join(dotfiles, snap.Dest)

		cmd := exec.Command("sh", "-c", snap.Command)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(w, "  %s  %s  %s\n",
				styleSyncWarn.Render("warn "),
				styleSyncPath.Render(snap.Dest),
				styleDim.Render("(command failed: "+err.Error()+")"),
			)
			if msg := strings.TrimSpace(stderr.String()); msg != "" {
				fmt.Fprintf(w, "         %s\n", styleDim.Render(msg))
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destAbs), 0o755); err != nil {
			fmt.Fprintf(w, "  %s  %s  %s\n",
				styleSyncWarn.Render("warn "),
				styleSyncPath.Render(snap.Dest),
				styleDim.Render("(cannot create dir: "+err.Error()+")"),
			)
			continue
		}

		if err := os.WriteFile(destAbs, stdout.Bytes(), 0o644); err != nil {
			fmt.Fprintf(w, "  %s  %s  %s\n",
				styleSyncWarn.Render("warn "),
				styleSyncPath.Render(snap.Dest),
				styleDim.Render("(cannot write: "+err.Error()+")"),
			)
			continue
		}

		fmt.Fprintf(w, "  %s  %s\n",
			styleSyncOK.Render("snap "),
			styleSyncPath.Render(snap.Dest),
		)
	}
}

// checkSymlinksNow logs warnings for any managed configs that are not correctly linked.
func checkSymlinksNow(w io.Writer, dotfiles string, manifest *SyncManifest) {
	if len(manifest.Configs) == 0 {
		fmt.Fprintf(w, "  %s\n", styleDim.Render("none configured"))
		return
	}

	allOK := true
	for _, sl := range checkSymlinks(dotfiles, manifest) {
		if sl.state == "correct" {
			continue
		}
		allOK = false
		fmt.Fprintf(w, "  %s  %s  %s  %s\n",
			styleSyncWarn.Render("warn "),
			styleSyncPath.Render(sl.target),
			styleDim.Render("→"),
			styleDim.Render(sl.state+": "+sl.source),
		)
	}
	if allOK {
		fmt.Fprintf(w, "  %s\n", styleSyncOK.Render("all symlinks correct"))
	}
}

// commitAndPush checks git status; if dirty, stages all, commits, and pushes.
// machine names the profile in the commit subject.
func commitAndPush(w io.Writer, dotfiles, machine string) error {
	// Check for changes
	statusCmd := exec.Command("git", "-C", dotfiles, "status", "--porcelain")
	var statusOut bytes.Buffer
	statusCmd.Stdout = &statusOut
	if err := statusCmd.Run(); err != nil {
		fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleDim.Render("git status failed: "+err.Error()))
		return fmt.Errorf("git status failed: %w", err)
	}

	if strings.TrimSpace(statusOut.String()) == "" {
		fmt.Fprintf(w, "  %s\n", styleSyncSkip.Render("nothing to commit"))
		return nil
	}

	// Build the subject before staging, while status still names the paths.
	msg := commitSubject(machine, parseStatusPaths(statusOut.String()))

	// Stage all changes
	addCmd := exec.Command("git", "-C", dotfiles, "add", "--all")
	var addOut bytes.Buffer
	addCmd.Stderr = &addOut
	if err := addCmd.Run(); err != nil {
		fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleDim.Render("git add failed: "+err.Error()))
		return fmt.Errorf("git add failed: %w", err)
	}

	// Commit
	commitCmd := exec.Command("git", "-C", dotfiles, "commit", "--no-verify", "-m", msg)
	var commitOut bytes.Buffer
	commitCmd.Stdout = &commitOut
	commitCmd.Stderr = &commitOut
	if err := commitCmd.Run(); err != nil {
		fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleDim.Render("git commit failed: "+commitOut.String()))
		return fmt.Errorf("git commit failed: %w", err)
	}
	fmt.Fprintf(w, "  %s  %s\n", styleSyncOK.Render("commit"), styleDim.Render(msg))

	// Push
	pushCmd := exec.Command("git", "-C", dotfiles, "push")
	var pushOut bytes.Buffer
	pushCmd.Stdout = &pushOut
	pushCmd.Stderr = &pushOut
	if err := pushCmd.Run(); err != nil {
		fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleDim.Render("git push failed: "+strings.TrimSpace(pushOut.String())))
		return fmt.Errorf("git push failed: %w", err)
	}
	fmt.Fprintf(w, "  %s\n", styleSyncOK.Render("pushed"))

	return nil
}
