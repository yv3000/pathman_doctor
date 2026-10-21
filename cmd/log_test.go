package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yv3000/pathman_doctor/internal"
)

func TestDiagnosticsAreLeveledAndSeparateFromConsole(t *testing.T) {
	store := newFakeStore()
	store.user = []string{filepath.Join(t.TempDir(), "missing")}
	app, out := newTestApp(t, store, "")
	var log bytes.Buffer
	app.Log = &log

	if code := app.Run([]string{"doctor"}); code != ExitOK {
		t.Fatalf("exit = %d, want %d", code, ExitOK)
	}

	diagnostics := log.String()
	for _, want := range []string{
		"[INFO] pathman doctor — scanning PATH entries...",
		"[INFO] Summary: 1 entries | 1 dead | 0 duplicate | action needed",
	} {
		if !strings.Contains(diagnostics, want) {
			t.Fatalf("diagnostics missing %q:\n%s", want, diagnostics)
		}
	}
	if strings.Contains(out.String(), "[INFO]") {
		t.Fatalf("console output contains a level tag: %q", out.String())
	}
	if !strings.Contains(out.String(), "DEAD") {
		t.Fatalf("console output missing the report: %q", out.String())
	}
}

func TestDiagnosticsTagWarningsAndErrors(t *testing.T) {
	store := newFakeStore()
	store.readErr[internal.ScopeSystem] = errors.New("access denied")
	store.readErr[internal.ScopeUser] = errors.New("access denied")
	app, _ := newTestApp(t, store, "")
	var log bytes.Buffer
	app.Log = &log

	if code := app.Run([]string{"fix"}); code != ExitFailure {
		t.Fatalf("exit = %d, want %d", code, ExitFailure)
	}

	diagnostics := log.String()
	for _, want := range []string{
		"[WARN] Could not read System PATH",
		"[ERROR] Could not read either PATH scope; nothing to fix.",
	} {
		if !strings.Contains(diagnostics, want) {
			t.Fatalf("diagnostics missing %q:\n%s", want, diagnostics)
		}
	}
}

func TestDiagnosticSinkIsOptIn(t *testing.T) {
	t.Setenv("PATHMAN_LOG", "")
	if got := diagnosticSink(); got != io.Discard {
		t.Fatalf("diagnosticSink() without PATHMAN_LOG = %#v, want io.Discard", got)
	}

	t.Setenv("PATHMAN_LOG", "1")
	if got := diagnosticSink(); got != os.Stderr {
		t.Fatalf("diagnosticSink() with PATHMAN_LOG = %#v, want os.Stderr", got)
	}
}
