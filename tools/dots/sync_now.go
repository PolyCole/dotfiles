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

// RunSyncNow performs the full sync cycle:
//  1. git pull --rebase
//  2. run each snapshot command and write output to dest
//  3. verify config symlinks (warn on broken)
//  4. if changes exist, commit and push; otherwise exit cleanly
func RunSyncNow(w io.Writer, dotfiles, machine string) error {
	if machine == "" {
		fmt.Fprintln(w, styleSyncErr.Render("error:"), styleDim.Render("DOTFILES_MACHINE is not set — cannot determine which sync.yml to use"))
		return fmt.Errorf("DOTFILES_MACHINE is not set")
	}

	fmt.Fprintf(w, "\n%s  %s\n\n",
		styleSyncLabel.Render("dots sync now"),
		styleDim.Render("— "+machine),
	)

	// ── Step 1: git pull --rebase ─────────────────────────────────────────
	fmt.Fprintf(w, "%s\n", styleGroupHeader.Render("pull"))
	if err := runGitPullRebase(w, dotfiles); err != nil {
		return err
	}

	// ── Step 2: load manifest ─────────────────────────────────────────────
	manifestPath := filepath.Join(dotfiles, "machines", machine, "sync.yml")
	manifest, err := LoadSyncManifest(manifestPath)
	if err != nil {
		fmt.Fprintln(w, styleSyncErr.Render("error:"), styleDim.Render(err.Error()))
		return err
	}

	// ── Step 3: run snapshots ─────────────────────────────────────────────
	fmt.Fprintf(w, "\n%s\n", styleGroupHeader.Render("snapshots"))
	runSnapshots(w, dotfiles, manifest)

	// ── Step 4: verify symlinks ───────────────────────────────────────────
	fmt.Fprintf(w, "\n%s\n", styleGroupHeader.Render("symlinks"))
	checkSymlinksNow(w, dotfiles, manifest)

	// ── Step 5: commit and push if dirty ─────────────────────────────────
	fmt.Fprintf(w, "\n%s\n", styleGroupHeader.Render("commit"))
	if err := commitAndPush(w, dotfiles); err != nil {
		return err
	}

	fmt.Fprintln(w)
	return nil
}

// runGitPullRebase runs git pull --rebase in dotfiles.
func runGitPullRebase(w io.Writer, dotfiles string) error {
	cmd := exec.Command("git", "-C", dotfiles, "pull", "--rebase")
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
func commitAndPush(w io.Writer, dotfiles string) error {
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

	// Stage all changes
	addCmd := exec.Command("git", "-C", dotfiles, "add", "--all")
	var addOut bytes.Buffer
	addCmd.Stderr = &addOut
	if err := addCmd.Run(); err != nil {
		fmt.Fprintf(w, "  %s  %s\n", styleSyncErr.Render("error"), styleDim.Render("git add failed: "+err.Error()))
		return fmt.Errorf("git add failed: %w", err)
	}

	// Commit
	msg := "sync: " + time.Now().Format("2006-01-02")
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
