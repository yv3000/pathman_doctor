package internal

import (
	"slices"
	"testing"
)

func testEntries() []Entry {
	return []Entry{
		{Scope: ScopeSystem, Path: `C:\Windows`, Expanded: `C:\Windows`, Status: StatusOK},
		{Scope: ScopeSystem, Path: `C:\dead`, Expanded: `C:\dead`, Status: StatusDead},
		{Scope: ScopeUser, Path: `C:\dup`, Expanded: `C:\dup`, Status: StatusDup},
		{Scope: ScopeUser, Path: `%USERPROFILE%\tools`, Expanded: `C:\Users\me\tools`, Status: StatusOK},
	}
}

func TestBuildPlanDefaultsToDeadAndDuplicate(t *testing.T) {
	plan := BuildPlan(testEntries(), Filter{})

	if len(plan.Remove) != 2 {
		t.Fatalf("removed %d entries, want 2", len(plan.Remove))
	}
	if !slices.Equal(plan.System, []string{`C:\Windows`}) {
		t.Fatalf("system entries = %#v", plan.System)
	}
	if !slices.Equal(plan.User, []string{`%USERPROFILE%\tools`}) {
		t.Fatalf("user entries = %#v", plan.User)
	}
	if got := plan.RemovedIn(ScopeSystem); got != 1 {
		t.Fatalf("RemovedIn(System) = %d, want 1", got)
	}
	if got := plan.RemovedIn(ScopeUser); got != 1 {
		t.Fatalf("RemovedIn(User) = %d, want 1", got)
	}
}

func TestBuildPlanOnlyDead(t *testing.T) {
	plan := BuildPlan(testEntries(), Filter{Only: StatusDead})

	if len(plan.Remove) != 1 || plan.Remove[0].Path != `C:\dead` {
		t.Fatalf("removed %#v, want only C:\\dead", plan.Remove)
	}
	if !slices.Equal(plan.User, []string{`C:\dup`, `%USERPROFILE%\tools`}) {
		t.Fatalf("user entries = %#v", plan.User)
	}
}

func TestBuildPlanOnlyDuplicate(t *testing.T) {
	plan := BuildPlan(testEntries(), Filter{Only: StatusDup})

	if len(plan.Remove) != 1 || plan.Remove[0].Path != `C:\dup` {
		t.Fatalf("removed %#v, want only C:\\dup", plan.Remove)
	}
}

func TestBuildPlanEntryFilterIsCaseInsensitive(t *testing.T) {
	entries := append(testEntries(), Entry{
		Scope:    ScopeUser,
		Path:     `%TEMP%\gone`,
		Expanded: `C:\Temp\gone`,
		Status:   StatusDead,
	})

	plan := BuildPlan(entries, Filter{Entry: `c:\dead\`})
	if len(plan.Remove) != 1 || plan.Remove[0].Path != `C:\dead` {
		t.Fatalf("removed %#v, want only C:\\dead", plan.Remove)
	}

	expanded := BuildPlan(entries, Filter{Entry: `C:\Temp\gone`})
	if len(expanded.Remove) != 1 || expanded.Remove[0].Path != `%TEMP%\gone` {
		t.Fatalf("removed %#v, want the %%TEMP%% entry", expanded.Remove)
	}
}

func TestBuildPlanKeepsHealthyEntries(t *testing.T) {
	plan := BuildPlan(testEntries()[0:1], Filter{})
	if len(plan.Remove) != 0 {
		t.Fatalf("removed %#v, want nothing", plan.Remove)
	}
	if !slices.Equal(plan.System, []string{`C:\Windows`}) {
		t.Fatalf("system entries = %#v", plan.System)
	}
}

func TestBuildPlanEntryFilterDoesNotRemoveHealthyEntry(t *testing.T) {
	plan := BuildPlan(testEntries(), Filter{Entry: `C:\Windows`})
	if len(plan.Remove) != 0 {
		t.Fatalf("removed %#v, want nothing", plan.Remove)
	}
}
