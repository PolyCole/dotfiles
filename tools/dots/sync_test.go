package dots

import (
	"os"
	"path/filepath"
	"testing"
)

func writeSyncManifest(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "sync.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeSyncManifest: %v", err)
	}
	return path
}

func TestLoadSyncManifest_FullManifest(t *testing.T) {
	path := writeSyncManifest(t, `
configs:
  - source: machines/personal/gitconfig
    target: ~/.gitconfig
  - source: zshrc
    target: ~/.zshrc

snapshots:
  - command: brew bundle dump --force --file=-
    dest: machines/personal/Brewfile
  - command: python -m pip list
    dest: machines/personal/python_packages
`)

	m, err := LoadSyncManifest(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(m.Configs) != 2 {
		t.Errorf("expected 2 configs, got %d", len(m.Configs))
	}
	if m.Configs[0].Source != "machines/personal/gitconfig" {
		t.Errorf("configs[0].Source = %q, want %q", m.Configs[0].Source, "machines/personal/gitconfig")
	}
	if m.Configs[0].Target != "~/.gitconfig" {
		t.Errorf("configs[0].Target = %q, want %q", m.Configs[0].Target, "~/.gitconfig")
	}
	if m.Configs[1].Source != "zshrc" {
		t.Errorf("configs[1].Source = %q, want %q", m.Configs[1].Source, "zshrc")
	}

	if len(m.Snapshots) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(m.Snapshots))
	}
	if m.Snapshots[0].Command != "brew bundle dump --force --file=-" {
		t.Errorf("snapshots[0].Command = %q, want %q", m.Snapshots[0].Command, "brew bundle dump --force --file=-")
	}
	if m.Snapshots[0].Dest != "machines/personal/Brewfile" {
		t.Errorf("snapshots[0].Dest = %q, want %q", m.Snapshots[0].Dest, "machines/personal/Brewfile")
	}
	if m.Snapshots[1].Dest != "machines/personal/python_packages" {
		t.Errorf("snapshots[1].Dest = %q, want %q", m.Snapshots[1].Dest, "machines/personal/python_packages")
	}
}

func TestLoadSyncManifest_ConfigsOnly(t *testing.T) {
	path := writeSyncManifest(t, `
configs:
  - source: zshrc
    target: ~/.zshrc
`)

	m, err := LoadSyncManifest(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Configs) != 1 {
		t.Errorf("expected 1 config, got %d", len(m.Configs))
	}
	if len(m.Snapshots) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(m.Snapshots))
	}
}

func TestLoadSyncManifest_SnapshotsOnly(t *testing.T) {
	path := writeSyncManifest(t, `
snapshots:
  - command: brew bundle dump --force --file=-
    dest: Brewfile
`)

	m, err := LoadSyncManifest(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.Configs) != 0 {
		t.Errorf("expected 0 configs, got %d", len(m.Configs))
	}
	if len(m.Snapshots) != 1 {
		t.Errorf("expected 1 snapshot, got %d", len(m.Snapshots))
	}
}

func TestLoadSyncManifest_MissingFile(t *testing.T) {
	_, err := LoadSyncManifest("/nonexistent/path/sync.yml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadSyncManifest_MalformedYAML(t *testing.T) {
	path := writeSyncManifest(t, `configs: [invalid yaml: {`)
	_, err := LoadSyncManifest(path)
	if err == nil {
		t.Fatal("expected error for malformed YAML, got nil")
	}
}

func TestLoadSyncManifest_EmptyFile(t *testing.T) {
	path := writeSyncManifest(t, ``)
	m, err := LoadSyncManifest(path)
	if err != nil {
		t.Fatalf("unexpected error for empty file: %v", err)
	}
	if len(m.Configs) != 0 {
		t.Errorf("expected 0 configs, got %d", len(m.Configs))
	}
	if len(m.Snapshots) != 0 {
		t.Errorf("expected 0 snapshots, got %d", len(m.Snapshots))
	}
}
