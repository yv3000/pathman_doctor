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
	app.RemoveAll = func(string) error { return errors.New("access is denied") }
	var scheduled []string
	app.ScheduleRemoval = func(dir string) error {
		scheduled = append(scheduled, dir)
		return nil
	}

	if code := app.Run([]string{"uninstall"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}
	if !slices.Equal(scheduled, []string{installDir}) {
		t.Fatalf("scheduled = %#v, want %q", scheduled, installDir)
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

func TestUninstallRejectsExtraArguments(t *testing.T) {
	app, _ := newTestApp(t, newFakeStore(), "y\n")

	if code := app.Run([]string{"uninstall", "now"}); code != ExitUsage {
		t.Fatalf("exit = %d, want %d", code, ExitUsage)
	}
}
