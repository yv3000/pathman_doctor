package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/yv3000/pathman_doctor/internal"
)

func TestMain(m *testing.M) {
	color.NoColor = true
	os.Exit(m.Run())
}

type fakeStore struct {
	user       []string
	system     []string
	readErr    map[internal.Scope]error
	raw        map[internal.Scope]string
	writeErr   map[internal.Scope]error
	writes     map[internal.Scope][]string
	broadcasts int
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		readErr:  map[internal.Scope]error{},
		raw:      map[internal.Scope]string{},
		writeErr: map[internal.Scope]error{},
		writes:   map[internal.Scope][]string{},
	}
}

func (f *fakeStore) ReadPATH(scope internal.Scope) ([]string, error) {
	if err := f.readErr[scope]; err != nil {
		return nil, err
	}
	if scope == internal.ScopeUser {
		return f.user, nil
	}
	return f.system, nil
}

func (f *fakeStore) ReadRaw(scope internal.Scope) (string, error) {
	if value, ok := f.raw[scope]; ok {
		return value, nil
	}
	if err := f.readErr[scope]; err != nil {
		return "", err
	}
	if scope == internal.ScopeUser {
		return internal.JoinList(f.user), nil
	}
	return internal.JoinList(f.system), nil
}

func (f *fakeStore) WritePATH(scope internal.Scope, entries []string) error {
	if err := f.writeErr[scope]; err != nil {
		return err
	}
	f.writes[scope] = entries
	return nil
}

func (f *fakeStore) Broadcast() {
	f.broadcasts++
}

func newTestApp(t *testing.T, store internal.Store, input string) (App, *bytes.Buffer) {
	t.Helper()
	out := &bytes.Buffer{}
	app := NewApp(store, strings.NewReader(input), out, "9.9.9")
	app.BackupDir = t.TempDir()
	return app, out
}

func TestRunHelpAndVersion(t *testing.T) {
	app, out := newTestApp(t, newFakeStore(), "")

	if code := app.Run(nil); code != ExitOK {
		t.Fatalf("no args exit = %d, want %d", code, ExitOK)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("help output missing usage: %q", out.String())
	}

	out.Reset()
	if code := app.Run([]string{"--version"}); code != ExitOK {
		t.Fatalf("version exit = %d, want %d", code, ExitOK)
	}
	if !strings.Contains(out.String(), "pathman v9.9.9") {
		t.Fatalf("version output = %q", out.String())
	}
}

func TestRunUnknownCommandIsUsageError(t *testing.T) {
	app, out := newTestApp(t, newFakeStore(), "")

	if code := app.Run([]string{"bogus"}); code != ExitUsage {
		t.Fatalf("exit = %d, want %d", code, ExitUsage)
	}
	if !strings.Contains(out.String(), "Unknown command") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestDoctorReportsIssues(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing")

	store := newFakeStore()
	store.system = []string{dir}
	store.user = []string{missing}
	app, out := newTestApp(t, store, "")

	if code := app.Run([]string{"doctor"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}
	text := out.String()
	for _, want := range []string{"DEAD", missing, "Summary: 2 entries | 1 dead | 0 duplicate"} {
		if !strings.Contains(text, want) {
			t.Fatalf("output missing %q:\n%s", want, text)
		}
	}
}

func TestDoctorWarnsWhenSystemPathIsUnreadable(t *testing.T) {
	store := newFakeStore()
	store.readErr[internal.ScopeSystem] = errors.New("access denied")
	store.user = []string{filepath.Join(t.TempDir(), "missing")}
	app, out := newTestApp(t, store, "")

	if code := app.Run([]string{"doctor"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}
	if !strings.Contains(out.String(), "Could not read System PATH") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestFixWarnsWhenSystemPathIsUnreadable(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")

	store := newFakeStore()
	store.readErr[internal.ScopeSystem] = errors.New("access denied")
	store.user = []string{missing}
	app, out := newTestApp(t, store, "y\n")

	if code := app.Run([]string{"fix"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}
	if !strings.Contains(out.String(), "Could not read System PATH") {
		t.Fatalf("output = %q", out.String())
	}

	written, ok := store.writes[internal.ScopeUser]
	if !ok {
		t.Fatalf("user PATH was not cleaned: %#v", store.writes)
	}
	if len(written) != 0 {
		t.Fatalf("user PATH = %#v, want empty", written)
	}
	if _, ok := store.writes[internal.ScopeSystem]; ok {
		t.Fatalf("system PATH must not be written when it could not be read")
	}
}

func TestFixFailsWhenBothScopesAreUnreadable(t *testing.T) {
	store := newFakeStore()
	store.readErr[internal.ScopeSystem] = errors.New("access denied")
	store.readErr[internal.ScopeUser] = errors.New("access denied")
	app, out := newTestApp(t, store, "y\n")

	if code := app.Run([]string{"fix"}); code != ExitFailure {
		t.Fatalf("exit = %d, want %d", code, ExitFailure)
	}
	if !strings.Contains(out.String(), "Could not read either PATH scope") {
		t.Fatalf("output = %q", out.String())
	}
	if len(store.writes) != 0 {
		t.Fatalf("writes = %#v, want none", store.writes)
	}
}

func TestFixCancelDoesNotWrite(t *testing.T) {
	store := newFakeStore()
	store.user = []string{filepath.Join(t.TempDir(), "missing")}
	app, out := newTestApp(t, store, "n\n")

	if code := app.Run([]string{"fix"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}
	if len(store.writes) != 0 {
		t.Fatalf("writes = %#v, want none", store.writes)
	}
	if !strings.Contains(out.String(), "Cancelled.") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestFixRemovesDeadAndDuplicateEntries(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing")

	store := newFakeStore()
	store.system = []string{dir}
	store.user = []string{missing, dir}
	app, out := newTestApp(t, store, "y\n")

	if code := app.Run([]string{"fix"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}

	written, ok := store.writes[internal.ScopeUser]
	if !ok {
		t.Fatalf("user PATH was not written: %#v", store.writes)
	}
	if len(written) != 0 {
		t.Fatalf("user PATH = %#v, want empty", written)
	}
	if _, ok := store.writes[internal.ScopeSystem]; ok {
		t.Fatalf("system PATH should not be written: %#v", store.writes)
	}
	if store.broadcasts != 1 {
		t.Fatalf("broadcasts = %d, want 1", store.broadcasts)
	}
	if !strings.Contains(out.String(), "2 entries removed") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestFixOnlyDuplicateKeepsDeadEntries(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing")

	store := newFakeStore()
	store.system = []string{dir}
	store.user = []string{missing, dir}
	app, _ := newTestApp(t, store, "y\n")

	if code := app.Run([]string{"fix", "--only", "dup"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}

	written := store.writes[internal.ScopeUser]
	if len(written) != 1 || written[0] != missing {
		t.Fatalf("user PATH = %#v, want %q", written, missing)
	}
}

func TestFixWritesBackupBeforeChange(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	raw := `C:\Tools;` + missing

	store := newFakeStore()
	store.user = []string{missing}
	store.raw[internal.ScopeUser] = raw
	app, _ := newTestApp(t, store, "y\n")

	if code := app.Run([]string{"fix"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}

	files, err := os.ReadDir(app.BackupDir)
	if err != nil {
		t.Fatalf("reading backup dir: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("backup files = %d, want 1", len(files))
	}

	content, err := os.ReadFile(filepath.Join(app.BackupDir, files[0].Name()))
	if err != nil {
		t.Fatalf("reading backup: %v", err)
	}
	if string(content) != raw {
		t.Fatalf("backup = %q, want %q", string(content), raw)
	}
}

func TestFixAbortsWhenBackupFails(t *testing.T) {
	store := newFakeStore()
	store.user = []string{filepath.Join(t.TempDir(), "missing")}

	file := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(file, []byte("blocker"), 0o600); err != nil {
		t.Fatalf("writing blocker file: %v", err)
	}

	app, out := newTestApp(t, store, "y\n")
	app.BackupDir = file

	if code := app.Run([]string{"fix"}); code != ExitFailure {
		t.Fatalf("exit = %d, want %d", code, ExitFailure)
	}
	if len(store.writes) != 0 {
		t.Fatalf("writes = %#v, want none", store.writes)
	}
	if !strings.Contains(out.String(), "Could not back up PATH") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestFixReportsPartialWriteFailure(t *testing.T) {
	dir := t.TempDir()
	systemMissing := filepath.Join(dir, "system-missing")
	userMissing := filepath.Join(dir, "user-missing")

	store := newFakeStore()
	store.system = []string{systemMissing}
	store.user = []string{userMissing}
	store.writeErr[internal.ScopeSystem] = errors.New("access denied")
	app, out := newTestApp(t, store, "y\n")

	if code := app.Run([]string{"fix"}); code != ExitFailure {
		t.Fatalf("exit = %d, want %d", code, ExitFailure)
	}
	if _, ok := store.writes[internal.ScopeUser]; !ok {
		t.Fatalf("user PATH should still be cleaned: %#v", store.writes)
	}
	if store.broadcasts != 1 {
		t.Fatalf("broadcasts = %d, want 1", store.broadcasts)
	}
	if !strings.Contains(out.String(), "System PATH was left unchanged") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestFixReportsUserWriteFailure(t *testing.T) {
	store := newFakeStore()
	store.user = []string{filepath.Join(t.TempDir(), "missing")}
	store.writeErr[internal.ScopeUser] = errors.New("access denied")
	app, out := newTestApp(t, store, "y\n")

	if code := app.Run([]string{"fix"}); code != ExitFailure {
		t.Fatalf("exit = %d, want %d", code, ExitFailure)
	}
	if !strings.Contains(out.String(), "Failed to write User PATH") {
		t.Fatalf("output = %q", out.String())
	}
	if !strings.Contains(out.String(), "No changes applied") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestFixEntryFilterTargetsOnePath(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "first")
	second := filepath.Join(dir, "second")

	store := newFakeStore()
	store.user = []string{first, second}
	app, _ := newTestApp(t, store, "y\n")

	if code := app.Run([]string{"fix", "--entry", second}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}

	written := store.writes[internal.ScopeUser]
	if len(written) != 1 || written[0] != first {
		t.Fatalf("user PATH = %#v, want %q", written, first)
	}
}

func TestFixRejectsInvalidOnlyValue(t *testing.T) {
	store := newFakeStore()
	store.user = []string{filepath.Join(t.TempDir(), "missing")}
	app, out := newTestApp(t, store, "y\n")

	if code := app.Run([]string{"fix", "--only", "bogus"}); code != ExitUsage {
		t.Fatalf("exit = %d, want %d", code, ExitUsage)
	}
	if len(store.writes) != 0 {
		t.Fatalf("writes = %#v, want none", store.writes)
	}
	if !strings.Contains(out.String(), "invalid --only value") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestFixRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing entry value", args: []string{"fix", "--entry"}},
		{name: "unknown flag", args: []string{"fix", "--nope"}},
		{name: "positional argument", args: []string{"fix", "extra"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeStore()
			store.user = []string{filepath.Join(t.TempDir(), "missing")}
			app, _ := newTestApp(t, store, "y\n")

			if code := app.Run(tt.args); code != ExitUsage {
				t.Fatalf("exit = %d, want %d", code, ExitUsage)
			}
			if len(store.writes) != 0 {
				t.Fatalf("writes = %#v, want none", store.writes)
			}
		})
	}
}

func TestFixHelpDoesNotTouchPATH(t *testing.T) {
	store := newFakeStore()
	store.user = []string{filepath.Join(t.TempDir(), "missing")}
	app, out := newTestApp(t, store, "")

	if code := app.Run([]string{"fix", "--help"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}
	if !strings.Contains(out.String(), "Usage: pathman fix") {
		t.Fatalf("output = %q", out.String())
	}
	if len(store.writes) != 0 {
		t.Fatalf("writes = %#v, want none", store.writes)
	}
}
