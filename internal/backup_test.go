package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveBackupWritesRawValue(t *testing.T) {
	dir := t.TempDir()
	raw := `C:\Windows;%USERPROFILE%\bin`

	path, err := SaveBackup(dir, ScopeUser, raw)
	if err != nil {
		t.Fatalf("SaveBackup() = %v", err)
	}
	if filepath.Dir(path) != dir {
		t.Fatalf("backup path = %q, want it inside %q", path, dir)
	}
	if !strings.Contains(filepath.Base(path), "User") {
		t.Fatalf("backup name = %q, want the scope in the name", filepath.Base(path))
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading backup: %v", err)
	}
	if string(got) != raw {
		t.Fatalf("backup content = %q, want %q", string(got), raw)
	}
}

func TestSaveBackupRequiresDirectory(t *testing.T) {
	if _, err := SaveBackup("", ScopeUser, `C:\Windows`); err == nil {
		t.Fatal("SaveBackup with an empty directory should fail")
	}
}

func TestDefaultBackupDirFallsBackToTempDir(t *testing.T) {
	t.Setenv("LOCALAPPDATA", "")
	want := filepath.Join(os.TempDir(), "pathman", "backups")
	if got := DefaultBackupDir(); got != want {
		t.Fatalf("DefaultBackupDir() = %q, want %q", got, want)
	}
}

func TestSaveBackupFailsWhenTargetIsNotADirectory(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("writing blocker file: %v", err)
	}

	if _, err := SaveBackup(blocker, ScopeUser, `C:\Windows`); err == nil {
		t.Fatal("SaveBackup into a file should fail")
	}
}

func TestDefaultBackupDirUsesLocalAppData(t *testing.T) {
	t.Setenv("LOCALAPPDATA", `C:\Users\me\AppData\Local`)
	want := filepath.Join(`C:\Users\me\AppData\Local`, "pathman", "backups")
	if got := DefaultBackupDir(); got != want {
		t.Fatalf("DefaultBackupDir() = %q, want %q", got, want)
	}
}
