package dots

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// syncResultPrefix starts the last line of every run in the log, so
// 'dots sync status' and 'dots doctor' can report how the last run ended.
const syncResultPrefix = "result: "

// pullRetryDelays spaces out pull retries when GitHub is unreachable.
// launchd runs a missed 09:00 job the moment the laptop wakes, usually
// before Wi-Fi or VPN is back, so the first attempt often fails.
var pullRetryDelays = []time.Duration{15 * time.Second, 30 * time.Second, 45 * time.Second, 60 * time.Second}

// errOffline marks a run skipped because GitHub stayed unreachable.
// It is not a failure: the next scheduled run catches up.
var errOffline = errors.New("github unreachable")

// RunSyncNow performs the full sync cycle and sends a desktop notification
// on failure so unattended launchd runs don't fail silently.
func RunSyncNow(w io.Writer, dotfiles, machine string) error {
	err := runSyncNow(w, dotfiles, machine)
	switch {
	case errors.Is(err, errOffline):
		fmt.Fprintf(w, "\n%s\n\n", styleSyncSkip.Render(syncResultPrefix+"skipped (offline)"))
		return nil
	case err != nil:
		fmt.Fprintf(w, "\n%s\n\n", styleSyncErr.Render(syncResultPrefix+"failed: "+err.Error()))
		notifySyncFailure(err.Error())
		return err
	}
	fmt.Fprintf(w, "\n%s\n\n", styleSyncOK.Render(syncResultPrefix+"ok"))
	return nil
}

// runSyncNow is the sync cycle itself:
//  1. git pull --rebase
//  2. rebuild binaries if the pull brought new commits
//  3. run each snapshot command and write output to dest
//  4. verify config symlinks (warn on broken)
//  5. commit if dirty, then push if ahead of upstream
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

	// ── Step 6: commit if dirty, push if ahead ───────────────────────────
	fmt.Fprintf(w, "\n%s\n", styleGroupHeader.Render("commit"))
	if err := commitIfDirty(w, dotfiles, machine); err != nil {
		return err
	}
	fmt.Fprintf(w, "\n%s\n", styleGroupHeader.Render("push"))
	return pushIfAhead(w, dotfiles)
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
// Network failures are retried per pullRetryDelays; if GitHub is still
// unreachable after the last retry, it returns errOffline.
func runGitPullRebase(w io.Writer, dotfiles string) error {
	for attempt := 0; ; attempt++ {
		cmd := exec.Command("git", "-C", dotfiles, "pull", "--rebase", "--autostash")
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		err := cmd.Run()
		if err == nil {
			printPullResult(w, strings.TrimSpace(out.String()))
			return nil
		}
		if !isNetworkError(out.String()) {
			fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleDim.Render(out.String()))
			return fmt.Errorf("git pull --rebase failed: %w", err)
		}
		if attempt == len(pullRetryDelays) {
			fmt.Fprintf(w, "  %s  %s\n", styleSyncSkip.Render("offline"),
				styleDim.Render(fmt.Sprintf("github unreachable after %d attempts — skipping this run", attempt+1)))
			return errOffline
		}
		fmt.Fprintf(w, "  %s  %s\n", styleSyncWarn.Render("retry"),
			styleDim.Render(fmt.Sprintf("github unreachable — retrying in %s", pullRetryDelays[attempt])))
		time.Sleep(pullRetryDelays[attempt])
	}
}

func printPullResult(w io.Writer, outStr string) {
	if outStr == "Current branch main is up to date." || strings.Contains(outStr, "up to date") {
		fmt.Fprintf(w, "  %s\n", styleSyncSkip.Render("already up to date"))
	} else if outStr != "" {
		fmt.Fprintf(w, "  %s\n", styleDim.Render(outStr))
	} else {
		fmt.Fprintf(w, "  %s\n", styleSyncOK.Render("pulled"))
	}
}

// networkErrorMarkers are git/ssh messages meaning GitHub couldn't be reached
// at all, as opposed to auth or merge failures, which should not be retried.
var networkErrorMarkers = []string{
	"Could not resolve hostname",
	"Could not resolve host",
	"connect to host",
	"Network is unreachable",
	"Operation timed out",
	"Connection timed out",
	"Connection reset",
}

func isNetworkError(out string) bool {
	for _, m := range networkErrorMarkers {
		if strings.Contains(out, m) {
			return true
		}
	}
	return false
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

// commitIfDirty checks git status; if dirty, stages all and commits.
// machine names the profile in the commit subject.
func commitIfDirty(w io.Writer, dotfiles, machine string) error {
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
	return nil
}

// pushIfAhead pushes whenever HEAD has commits upstream lacks — not only
// after this run committed — so a commit stranded by an earlier failed push
// goes out on the next run. A rejected push (another machine pushed first)
// gets one pull --rebase and a retry.
func pushIfAhead(w io.Writer, dotfiles string) error {
	ahead, _, err := aheadBehind(dotfiles)
	if err != nil {
		fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleDim.Render("cannot compare with upstream: "+err.Error()))
		return fmt.Errorf("cannot compare with upstream: %w", err)
	}
	if ahead == 0 {
		fmt.Fprintf(w, "  %s\n", styleSyncSkip.Render("nothing to push"))
		return nil
	}

	out, err := gitPush(dotfiles)
	if err != nil && strings.Contains(out, "[rejected]") {
		fmt.Fprintf(w, "  %s  %s\n", styleSyncWarn.Render("retry"), styleDim.Render("push rejected — remote moved; rebasing and retrying"))
		headBefore := gitHead(dotfiles)
		if err := runGitPullRebase(w, dotfiles); err != nil {
			return err
		}
		rebuildIfChanged(w, dotfiles, headBefore)
		out, err = gitPush(dotfiles)
	}
	if err != nil {
		fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleDim.Render("git push failed: "+out))
		return fmt.Errorf("git push failed: %w", err)
	}
	fmt.Fprintf(w, "  %s  %s\n", styleSyncOK.Render("pushed"), styleDim.Render(fmt.Sprintf("%d commit(s)", ahead)))
	return nil
}

func gitPush(dotfiles string) (string, error) {
	cmd := exec.Command("git", "-C", dotfiles, "push")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}
