package dots

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ── formatAge ──────────────────────────────────────────────────────────────

func TestFormatAge_JustNow(t *testing.T) {
	got := formatAge(time.Now().Add(-30 * time.Second))
	if got != "just now" {
		t.Errorf("expected 'just now', got %q", got)
	}
}

func TestFormatAge_Minutes(t *testing.T) {
	got := formatAge(time.Now().Add(-5 * time.Minute))
	if !strings.Contains(got, "minute") {
		t.Errorf("expected minutes string, got %q", got)
	}
}

func TestFormatAge_Hours(t *testing.T) {
	got := formatAge(time.Now().Add(-3 * time.Hour))
	if !strings.Contains(got, "hour") {
		t.Errorf("expected hours string, got %q", got)
	}
}

func TestFormatAge_Days(t *testing.T) {
	got := formatAge(time.Now().Add(-48 * time.Hour))
	if !strings.Contains(got, "day") {
		t.Errorf("expected days string, got %q", got)
	}
}

func TestFormatAge_SingularMinute(t *testing.T) {
	got := formatAge(time.Now().Add(-1 * time.Minute))
	if got != "1 minute ago" {
		t.Errorf("expected '1 minute ago', got %q", got)
	}
}

func TestFormatAge_SingularHour(t *testing.T) {
	got := formatAge(time.Now().Add(-1 * time.Hour))
	if got != "1 hour ago" {
		t.Errorf("expected '1 hour ago', got %q", got)
	}
}

func TestFormatAge_SingularDay(t *testing.T) {
	got := formatAge(time.Now().Add(-24 * time.Hour))
	if got != "1 day ago" {
		t.Errorf("expected '1 day ago', got %q", got)
	}
}

// ── checkSymlinks ──────────────────────────────────────────────────────────

func TestCheckSymlinks_Correct(t *testing.T) {
	dotfiles := t.TempDir()
	// Create source file
	srcRel := "test-config"
	srcAbs := filepath.Join(dotfiles, srcRel)
	if err := os.WriteFile(srcAbs, []byte("config"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create target directory and correct symlink
	tgtDir := t.TempDir()
	tgtAbs := filepath.Join(tgtDir, "config")
	if err := os.Symlink(srcAbs, tgtAbs); err != nil {
		t.Fatal(err)
	}

	manifest := &SyncManifest{
		Configs: []SyncConfig{{Source: srcRel, Target: tgtAbs}},
	}
	results := checkSymlinks(dotfiles, manifest)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].state != "correct" {
		t.Errorf("expected 'correct', got %q", results[0].state)
	}
}

func TestCheckSymlinks_Missing(t *testing.T) {
	dotfiles := t.TempDir()
	srcRel := "test-config"
	srcAbs := filepath.Join(dotfiles, srcRel)
	if err := os.WriteFile(srcAbs, []byte("config"), 0o644); err != nil {
		t.Fatal(err)
	}

	tgtDir := t.TempDir()
	tgtAbs := filepath.Join(tgtDir, "nonexistent-config")

	manifest := &SyncManifest{
		Configs: []SyncConfig{{Source: srcRel, Target: tgtAbs}},
	}
	results := checkSymlinks(dotfiles, manifest)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].state != "missing" {
		t.Errorf("expected 'missing', got %q", results[0].state)
	}
}

func TestCheckSymlinks_Conflict(t *testing.T) {
	dotfiles := t.TempDir()
	srcRel := "test-config"
	srcAbs := filepath.Join(dotfiles, srcRel)
	if err := os.WriteFile(srcAbs, []byte("config"), 0o644); err != nil {
		t.Fatal(err)
	}

	tgtDir := t.TempDir()
	tgtAbs := filepath.Join(tgtDir, "config")
	// Regular file at target — should be a conflict
	if err := os.WriteFile(tgtAbs, []byte("other"), 0o644); err != nil {
		t.Fatal(err)
	}

	manifest := &SyncManifest{
		Configs: []SyncConfig{{Source: srcRel, Target: tgtAbs}},
	}
	results := checkSymlinks(dotfiles, manifest)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].state != "conflict" {
		t.Errorf("expected 'conflict', got %q", results[0].state)
	}
}

func TestCheckSymlinks_Broken(t *testing.T) {
	dotfiles := t.TempDir()
	srcRel := "test-config"

	tgtDir := t.TempDir()
	tgtAbs := filepath.Join(tgtDir, "config")
	// Symlink pointing to non-existent source
	if err := os.Symlink("/nonexistent/path/nowhere", tgtAbs); err != nil {
		t.Fatal(err)
	}

	manifest := &SyncManifest{
		Configs: []SyncConfig{{Source: srcRel, Target: tgtAbs}},
	}
	results := checkSymlinks(dotfiles, manifest)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].state != "broken" {
		t.Errorf("expected 'broken', got %q", results[0].state)
	}
}

func TestCheckSymlinks_WrongSymlink(t *testing.T) {
	dotfiles := t.TempDir()
	srcRel := "test-config"
	srcAbs := filepath.Join(dotfiles, srcRel)
	if err := os.WriteFile(srcAbs, []byte("config"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Another file the symlink currently points to
	otherAbs := filepath.Join(dotfiles, "other-config")
	if err := os.WriteFile(otherAbs, []byte("other"), 0o644); err != nil {
		t.Fatal(err)
	}

	tgtDir := t.TempDir()
	tgtAbs := filepath.Join(tgtDir, "config")
	if err := os.Symlink(otherAbs, tgtAbs); err != nil {
		t.Fatal(err)
	}

	manifest := &SyncManifest{
		Configs: []SyncConfig{{Source: srcRel, Target: tgtAbs}},
	}
	results := checkSymlinks(dotfiles, manifest)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].state != "conflict" {
		t.Errorf("expected 'conflict' for wrong symlink target, got %q", results[0].state)
	}
}

// ── RunSyncStatus smoke test ───────────────────────────────────────────────

func TestRunSyncStatus_NoMachine(t *testing.T) {
	dotfiles := t.TempDir()
	var sb strings.Builder
	// Should not return an error even with empty machine
	if err := RunSyncStatus(&sb, dotfiles, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, "dots sync status") {
		t.Error("output should contain 'dots sync status'")
	}
	if !strings.Contains(out, "launchd job") {
		t.Error("output should contain launchd section")
	}
	// When machine is empty, configs/snapshots sections should be omitted
	if strings.Contains(out, "managed configs") {
		t.Error("output should not contain managed configs section when machine is empty")
	}
}

func TestRunSyncStatus_WithManifest(t *testing.T) {
	dotfiles := t.TempDir()

	// Write a minimal sync.yml
	machineDir := filepath.Join(dotfiles, "machines", "testmachine")
	if err := os.MkdirAll(machineDir, 0o755); err != nil {
		t.Fatal(err)
	}
	syncYML := `
configs:
  - source: zshrc
    target: /tmp/dots-test-zshrc-nonexistent
snapshots:
  - command: echo hello
    dest: machines/testmachine/hello.txt
`
	if err := os.WriteFile(filepath.Join(machineDir, "sync.yml"), []byte(syncYML), 0o644); err != nil {
		t.Fatal(err)
	}

	var sb strings.Builder
	if err := RunSyncStatus(&sb, dotfiles, "testmachine"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()

	if !strings.Contains(out, "managed configs") {
		t.Error("output should contain 'managed configs' section")
	}
	if !strings.Contains(out, "snapshots") {
		t.Error("output should contain 'snapshots' section")
	}
	// The zshrc target doesn't exist, should show as missing
	if !strings.Contains(out, "missing") {
		t.Error("output should show 'missing' for nonexistent symlink target")
	}
	// The snapshot has no git history, should show 'never captured'
	if !strings.Contains(out, "never captured") {
		t.Error("output should show 'never captured' for snapshot with no git history")
	}
}

func TestRunSyncStatus_NotInstalledMessage(t *testing.T) {
	// This test verifies the "not installed" path. Since we're not running
	// as a system user with a real LaunchAgents dir, the plist won't exist.
	dotfiles := t.TempDir()
	var sb strings.Builder
	_ = RunSyncStatus(&sb, dotfiles, "")
	out := sb.String()

	// Either "not installed" or "installed" should appear (depending on the machine)
	if !strings.Contains(out, "installed") && !strings.Contains(out, "loaded") {
		t.Error("output should contain some launchd status ('installed', 'not installed', or 'loaded')")
	}
}
