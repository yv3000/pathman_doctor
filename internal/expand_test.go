package internal

import (
	"path/filepath"
	"testing"
)

func TestExpandResolvesWindowsPercentVariables(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "tools")
	t.Setenv("PATHMAN_TEST_DIR", dir)

	if got, want := Expand(`%PATHMAN_TEST_DIR%\bin`), dir+`\bin`; got != want {
		t.Fatalf("Expand percent variable = %q, want %q", got, want)
	}
}

func TestExpandLeavesUnknownVariablesUntouched(t *testing.T) {
	const raw = `%PATHMAN_UNDEFINED_VAR%\bin`
	if got := Expand(raw); got != raw {
		t.Fatalf("Expand unknown variable = %q, want %q", got, raw)
	}
}

func TestExpandResolvesDollarVariables(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "tools")
	t.Setenv("PATHMAN_TEST_DOLLAR", dir)

	if got, want := Expand(`${PATHMAN_TEST_DOLLAR}/bin`), dir+`/bin`; got != want {
		t.Fatalf("Expand dollar variable = %q, want %q", got, want)
	}
}

func TestAnalyzePATHExpandsBeforeCheckingExistence(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATHMAN_TEST_DIR", dir)

	entries := AnalyzePATH(nil, []string{`%PATHMAN_TEST_DIR%`})
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].Status != StatusOK {
		t.Fatalf("status = %s, want OK (entry %q, expanded %q)", entries[0].Status, entries[0].Path, entries[0].Expanded)
	}
}
