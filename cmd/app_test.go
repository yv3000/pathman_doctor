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

func newTestApp(store internal.Store, input string) (App, *bytes.Buffer) {
	out := &bytes.Buffer{}
	return NewApp(store, strings.NewReader(input), out, "9.9.9"), out
}

func TestRunHelpAndVersion(t *testing.T) {
	app, out := newTestApp(newFakeStore(), "")

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
	app, out := newTestApp(newFakeStore(), "")

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
	app, out := newTestApp(store, "")

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
	app, out := newTestApp(store, "")

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
	app, out := newTestApp(store, "y\n")

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
	app, out := newTestApp(store, "y\n")

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
	app, out := newTestApp(store, "n\n")

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
	app, out := newTestApp(store, "y\n")

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
	app, _ := newTestApp(store, "y\n")

	if code := app.Run([]string{"fix", "--only", "dup"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}

	written := store.writes[internal.ScopeUser]
	if len(written) != 1 || written[0] != missing {
		t.Fatalf("user PATH = %#v, want %q", written, missing)
	}
}

func TestFixEntryFilterTargetsOnePath(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "first")
	second := filepath.Join(dir, "second")

	store := newFakeStore()
	store.user = []string{first, second}
	app, _ := newTestApp(store, "y\n")

	if code := app.Run([]string{"fix", "--entry", second}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}

	written := store.writes[internal.ScopeUser]
	if len(written) != 1 || written[0] != first {
		t.Fatalf("user PATH = %#v, want %q", written, first)
	}
}

func TestFixInvalidOnlyValueRemovesNothing(t *testing.T) {
	store := newFakeStore()
	store.user = []string{filepath.Join(t.TempDir(), "missing")}
	app, out := newTestApp(store, "y\n")

	if code := app.Run([]string{"fix", "--only", "bogus"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}
	if len(store.writes) != 0 {
		t.Fatalf("writes = %#v, want none", store.writes)
	}
	if !strings.Contains(out.String(), "Nothing to fix") {
		t.Fatalf("output = %q", out.String())
	}
}
