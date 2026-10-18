package internal

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestSplitList(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{name: "empty", in: "", want: nil},
		{name: "single entry", in: `C:\Windows`, want: []string{`C:\Windows`}},
		{name: "trailing separator", in: `C:\one;C:\two;`, want: []string{`C:\one`, `C:\two`}},
		{name: "blank segments", in: `; C:\one ;; ;C:\two`, want: []string{`C:\one`, `C:\two`}},
		{name: "trims whitespace", in: "  C:\\one  ;\tC:\\two\t", want: []string{`C:\one`, `C:\two`}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitList(tt.in)
			if !slices.Equal(got, tt.want) {
				t.Fatalf("SplitList(%q) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}

func TestJoinListRoundTrip(t *testing.T) {
	value := `C:\Windows;C:\Windows\System32;%USERPROFILE%\bin`
	if got := JoinList(SplitList(value)); got != value {
		t.Fatalf("round trip = %q, want %q", got, value)
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "whitespace only", in: "   ", want: ""},
		{name: "lowercases", in: `C:\Windows`, want: `c:\windows`},
		{name: "trailing separator", in: `C:\Windows\`, want: `c:\windows`},
		{name: "relative segments", in: `C:\Windows\System32\..`, want: `c:\windows`},
		{name: "trims space", in: `  C:\Windows  `, want: `c:\windows`},
		{name: "root stays intact", in: `C:\`, want: `c:\`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Normalize(tt.in); got != tt.want {
				t.Fatalf("Normalize(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestAnalyzeClassifiesEntries(t *testing.T) {
	analyzer := Analyzer{
		Expand: func(value string) string { return value },
		Exists: func(path string) bool { return path != `C:\missing` },
	}

	entries := analyzer.Analyze(
		[]string{`C:\tools`, `C:\missing`},
		[]string{`C:\Windows`, `C:\tools`, `C:\tools`},
	)

	if len(entries) != 5 {
		t.Fatalf("got %d entries, want 5", len(entries))
	}

	want := []struct {
		scope  Scope
		status Status
	}{
		{ScopeSystem, StatusOK},
		{ScopeSystem, StatusOK},
		{ScopeSystem, StatusDup},
		{ScopeUser, StatusDup},
		{ScopeUser, StatusDead},
	}
	for i, tt := range want {
		if entries[i].Scope != tt.scope || entries[i].Status != tt.status {
			t.Fatalf("entry %d = %s/%s, want %s/%s", i, entries[i].Scope, entries[i].Status, tt.scope, tt.status)
		}
	}

	if note := entries[2].Note; note != "(duplicate — appears 3x)" {
		t.Fatalf("duplicate note = %q", note)
	}
	if note := entries[4].Note; note != "(folder missing)" {
		t.Fatalf("dead note = %q", note)
	}
}

func TestAnalyzeIsCaseInsensitiveForDuplicates(t *testing.T) {
	analyzer := Analyzer{
		Expand: func(value string) string { return value },
		Exists: func(string) bool { return true },
	}

	entries := analyzer.Analyze(
		[]string{`c:\windows\`},
		[]string{`C:\Windows`},
	)

	if entries[0].Status != StatusOK {
		t.Fatalf("system entry status = %s, want OK", entries[0].Status)
	}
	if entries[1].Status != StatusDup {
		t.Fatalf("user entry status = %s, want DUP", entries[1].Status)
	}
}

func TestAnalyzeCountsDeadDuplicateAsDuplicate(t *testing.T) {
	analyzer := Analyzer{
		Expand: func(value string) string { return value },
		Exists: func(string) bool { return false },
	}

	entries := analyzer.Analyze(nil, []string{`C:\gone`, `C:\gone`})
	if entries[0].Status != StatusDead {
		t.Fatalf("first entry status = %s, want DEAD", entries[0].Status)
	}
	if entries[1].Status != StatusDup {
		t.Fatalf("second entry status = %s, want DUP", entries[1].Status)
	}
}

func TestAnalyzePATHUsesFilesystem(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "does-not-exist")

	entries := AnalyzePATH([]string{dir}, []string{missing})

	if entries[0].Status != StatusDead {
		t.Fatalf("missing directory status = %s, want DEAD", entries[0].Status)
	}
	if entries[1].Status != StatusOK {
		t.Fatalf("directory status = %s, want OK", entries[1].Status)
	}
}
