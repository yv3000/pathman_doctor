package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/yv3000/pathman_doctor/internal"
)

func TestUninstallRemovesInstallDirAndPathEntry(t *testing.T) {
	installDir := filepath.Join(t.TempDir(), ".pathman", "bin")
	other := filepath.Join(t.TempDir(), "other")

	store := newFakeStore()
	store.user = []string{other, installDir}
	store.raw[internal.ScopeUser] = internal.JoinList(store.user)

	app, out := newTestApp(t, store, "y\n")
	app.Executable = func() (string, error) {
		return filepath.Join(installDir, "pathman.exe"), nil
	}
	var removed []string
	app.RemoveAll = func(path string) error {
		removed = append(removed, path)
		return nil
	}

	if code := app.Run([]string{"uninstall"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}
	if !slices.Equal(removed, []string{installDir}) {
		t.Fatalf("removed = %#v, want %q", removed, installDir)
	}
	written := store.writes[internal.ScopeUser]
	if !slices.Equal(written, []string{other}) {
		t.Fatalf("user PATH = %#v, want %q", written, other)
	}
	if store.broadcasts != 1 {
		t.Fatalf("broadcasts = %d, want 1", store.broadcasts)
	}

	files, err := os.ReadDir(app.BackupDir)
	if err != nil {
		t.Fatalf("reading backup dir: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("backup files = %d, want 1", len(files))
	}
	if !strings.Contains(out.String(), "uninstalled") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestUninstallSchedulesRemovalWhenDeleteFails(t *testing.T) {
	installDir := filepath.Join(t.TempDir(), ".pathman", "bin")

	store := newFakeStore()
	store.user = []string{installDir}
	app, out := newTestApp(t, store, "y\n")
	app.Executable = func() (string, error) {
		return filepath.Join(installDir, "pathman.exe"), nil
	}
	var order []string
	app.RemoveAll = func(string) error {
		order = append(order, "remove")
		return errors.New("access is denied")
	}
	var scheduled []string
	app.ScheduleRemoval = func(dir string) error {
		order = append(order, "schedule")
		scheduled = append(scheduled, dir)
		return nil
	}

	if code := app.Run([]string{"uninstall"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}
	if !slices.Equal(scheduled, []string{installDir}) {
		t.Fatalf("scheduled = %#v, want %q", scheduled, installDir)
	}
	if !slices.Equal(order, []string{"remove", "schedule"}) {
		t.Fatalf("cleanup order = %#v, want a removal attempt before scheduling", order)
	}
	if len(store.writes[internal.ScopeUser]) != 0 {
		t.Fatalf("user PATH = %#v, want empty", store.writes[internal.ScopeUser])
	}
	if !strings.Contains(out.String(), "will be deleted when it exits") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestUninstallFallsBackToManualHint(t *testing.T) {
	installDir := filepath.Join(t.TempDir(), ".pathman", "bin")

	store := newFakeStore()
	store.user = []string{installDir}
	app, out := newTestApp(t, store, "y\n")
	app.Executable = func() (string, error) {
		return filepath.Join(installDir, "pathman.exe"), nil
	}
	app.RemoveAll = func(string) error { return errors.New("access is denied") }
	app.ScheduleRemoval = func(string) error { return errors.New("blocked") }

	if code := app.Run([]string{"uninstall"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}
	if !strings.Contains(out.String(), "delete the folder manually") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestDefaultScheduleRemovalRejectsUnsafePath(t *testing.T) {
	if err := defaultScheduleRemoval(`C:\bad"path`); err == nil {
		t.Fatal("defaultScheduleRemoval should reject quotes in the path")
	}
}

func TestUninstallRefusesDirectoryItDoesNotOwn(t *testing.T) {
	app, out := newTestApp(t, newFakeStore(), "y\n")
	app.Executable = func() (string, error) {
		return filepath.Join(t.TempDir(), "pathman.exe"), nil
	}
	var removed []string
	app.RemoveAll = func(path string) error {
		removed = append(removed, path)
		return nil
	}

	if code := app.Run([]string{"uninstall"}); code != ExitFailure {
		t.Fatalf("exit = %d, want %d", code, ExitFailure)
	}
	if len(removed) != 0 {
		t.Fatalf("removed = %#v, want none", removed)
	}
	if !strings.Contains(out.String(), "nothing was removed") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestUninstallCancelDoesNothing(t *testing.T) {
	installDir := filepath.Join(t.TempDir(), ".pathman", "bin")

	store := newFakeStore()
	store.user = []string{installDir}
	app, _ := newTestApp(t, store, "n\n")
	app.Executable = func() (string, error) {
		return filepath.Join(installDir, "pathman.exe"), nil
	}
	var removed []string
	app.RemoveAll = func(path string) error {
		removed = append(removed, path)
		return nil
	}

	if code := app.Run([]string{"uninstall"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}
	if len(removed) != 0 || len(store.writes) != 0 {
		t.Fatalf("cancel removed files %#v or wrote %#v", removed, store.writes)
	}
}

func TestUninstallStopsWhenUserPATHIsUnreadable(t *testing.T) {
	installDir := filepath.Join(t.TempDir(), ".pathman", "bin")

	store := newFakeStore()
	store.readErr[internal.ScopeUser] = errors.New("access denied")

	app, _ := newTestApp(t, store, "y\n")
	app.Executable = func() (string, error) {
		return filepath.Join(installDir, "pathman.exe"), nil
	}
	var removed []string
	app.RemoveAll = func(path string) error {
		removed = append(removed, path)
		return nil
	}

	if code := app.Run([]string{"uninstall"}); code != ExitFailure {
		t.Fatalf("exit = %d, want %d", code, ExitFailure)
	}
	if len(removed) != 0 {
		t.Fatalf("removed = %#v, want none", removed)
	}
}

func TestUninstallSkipsPathWriteWhenEntryIsAbsent(t *testing.T) {
	installDir := filepath.Join(t.TempDir(), ".pathman", "bin")
	other := filepath.Join(t.TempDir(), "other")

	store := newFakeStore()
	store.user = []string{other}
	app, out := newTestApp(t, store, "y\n")
	app.Executable = func() (string, error) {
		return filepath.Join(installDir, "pathman.exe"), nil
	}
	var removed []string
	app.RemoveAll = func(path string) error {
		removed = append(removed, path)
		return nil
	}

	if code := app.Run([]string{"uninstall"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}
	if !slices.Equal(removed, []string{installDir}) {
		t.Fatalf("removed = %#v, want %q", removed, installDir)
	}
	if len(store.writes) != 0 {
		t.Fatalf("writes = %#v, want none when the entry is absent", store.writes)
	}
	if store.broadcasts != 0 {
		t.Fatalf("broadcasts = %d, want 0", store.broadcasts)
	}
	if !strings.Contains(out.String(), "uninstalled") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestUninstallLeavesFilesWhenPathWriteFails(t *testing.T) {
	installDir := filepath.Join(t.TempDir(), ".pathman", "bin")

	store := newFakeStore()
	store.user = []string{installDir}
	store.writeErr[internal.ScopeUser] = errors.New("access denied")
	app, out := newTestApp(t, store, "y\n")
	app.Executable = func() (string, error) {
		return filepath.Join(installDir, "pathman.exe"), nil
	}
	var removed []string
	app.RemoveAll = func(path string) error {
		removed = append(removed, path)
		return nil
	}

	if code := app.Run([]string{"uninstall"}); code != ExitFailure {
		t.Fatalf("exit = %d, want %d", code, ExitFailure)
	}
	if len(removed) != 0 {
		t.Fatalf("removed = %#v, want none after a failed PATH write", removed)
	}
	if !strings.Contains(out.String(), "Could not remove pathman from User PATH") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestIsInstallDir(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "install dir", path: `C:\Users\me\.pathman\bin`, want: true},
		{name: "case insensitive", path: `C:\Users\me\.PATHMAN\BIN`, want: true},
		{name: "parent only", path: `C:\Users\me\.pathman`, want: false},
		{name: "child of install dir", path: `C:\Users\me\.pathman\bin\extra`, want: false},
		{name: "unrelated bin", path: `C:\Users\me\bin`, want: false},
		{name: "other bin", path: `C:\Tools\bin`, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isInstallDir(tt.path); got != tt.want {
				t.Fatalf("isInstallDir(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestDefaultRemoveAllDeletesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".pathman", "bin")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("creating install dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pathman.exe"), []byte("stub"), 0o600); err != nil {
		t.Fatalf("writing binary: %v", err)
	}

	if err := defaultRemoveAll(dir); err != nil {
		t.Fatalf("defaultRemoveAll() = %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("install dir still exists: %v", err)
	}
}

func TestDefaultExecutableReturnsRunningBinary(t *testing.T) {
	path, err := defaultExecutable()
	if err != nil {
		t.Fatalf("defaultExecutable() = %v", err)
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("defaultExecutable() = %q, want an absolute path", path)
	}
}

func TestUninstallRejectsExtraArguments(t *testing.T) {
	app, _ := newTestApp(t, newFakeStore(), "y\n")

	if code := app.Run([]string{"uninstall", "now"}); code != ExitUsage {
		t.Fatalf("exit = %d, want %d", code, ExitUsage)
	}
}
