package cmd

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestConfirm(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "lowercase yes", input: "y\n", want: true},
		{name: "uppercase yes", input: " Y \n", want: true},
		{name: "no", input: "n\n", want: false},
		{name: "empty line", input: "\n", want: false},
		{name: "end of input", input: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, _ := newTestApp(t, newFakeStore(), tt.input)
			if got := app.confirm("Apply?"); got != tt.want {
				t.Fatalf("confirm(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestFixTreatsEOFAsCancellation proves a closed stdin cannot apply a fix: the
// prompt returns false, nothing is written and the command reports a cancel.
func TestFixTreatsEOFAsCancellation(t *testing.T) {
	store := newFakeStore()
	store.user = []string{filepath.Join(t.TempDir(), "missing")}
	app, out := newTestApp(t, store, "")

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
