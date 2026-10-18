package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Scope identifies which PATH environment variable an entry belongs to.
type Scope string

const (
	ScopeUser   Scope = "User"
	ScopeSystem Scope = "System"
)

// Status is the health of a single PATH entry.
type Status string

const (
	StatusOK   Status = "OK"
	StatusDead Status = "DEAD"
	StatusDup  Status = "DUP"
)

// Entry is a single PATH entry together with the result of analysing it.
type Entry struct {
	Scope    Scope
	Path     string // raw value exactly as stored in the registry
	Expanded string // value after environment variable expansion
	Status   Status
	Note     string // human readable detail, e.g. "(folder missing)"
}

// Analyzer expands and checks PATH entries. The function fields make the
// analysis deterministic in tests without touching the real filesystem.
type Analyzer struct {
	Expand func(string) string
	Exists func(string) bool
}

// DefaultAnalyzer analyses entries against the live environment.
func DefaultAnalyzer() Analyzer {
	return Analyzer{Expand: Expand, Exists: dirExists}
}

// Expand resolves environment variables inside a PATH entry.
func Expand(raw string) string {
	return os.ExpandEnv(raw)
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// SplitList turns a semicolon separated PATH value into entries, dropping
// empty and whitespace-only segments.
func SplitList(value string) []string {
	var entries []string
	for _, part := range strings.Split(value, ";") {
		part = strings.TrimSpace(part)
		if part != "" {
			entries = append(entries, part)
		}
	}
	return entries
}

// JoinList is the inverse of SplitList.
func JoinList(entries []string) string {
	return strings.Join(entries, ";")
}

// Normalize produces the key used to compare two PATH entries: case
// insensitive, cleaned, without a trailing separator.
func Normalize(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	return strings.ToLower(filepath.Clean(trimmed))
}

// Analyze classifies every entry as OK, DEAD or DUP. System entries are
// processed first, so a path that exists in both scopes keeps its System
// occurrence and the User one is reported as a duplicate.
func (a Analyzer) Analyze(userEntries, systemEntries []string) []Entry {
	all := make([]Entry, 0, len(systemEntries)+len(userEntries))
	counts := make(map[string]int, len(systemEntries)+len(userEntries))

	for _, raw := range systemEntries {
		expanded := a.Expand(raw)
		key := Normalize(expanded)
		counts[key]++
		all = append(all, Entry{Scope: ScopeSystem, Path: raw, Expanded: expanded})
	}
	for _, raw := range userEntries {
		expanded := a.Expand(raw)
		key := Normalize(expanded)
		counts[key]++
		all = append(all, Entry{Scope: ScopeUser, Path: raw, Expanded: expanded})
	}

	seen := make(map[string]bool, len(all))
	for i := range all {
		key := Normalize(all[i].Expanded)
		switch {
		case seen[key]:
			all[i].Status = StatusDup
			all[i].Note = fmt.Sprintf("(duplicate — appears %dx)", counts[key])
		case !a.Exists(all[i].Expanded):
			all[i].Status = StatusDead
			all[i].Note = "(folder missing)"
		default:
			all[i].Status = StatusOK
		}
		seen[key] = true
	}

	return all
}

// AnalyzePATH analyses entries using the live environment.
func AnalyzePATH(userEntries, systemEntries []string) []Entry {
	return DefaultAnalyzer().Analyze(userEntries, systemEntries)
}

// Filter narrows down which entries a fix should remove.
type Filter struct {
	Only  Status // empty means DEAD and DUP
	Entry string // compare against this path only, when set
}

// Plan is the outcome of applying a Filter: the entries to remove and the
// PATH values that remain for each scope.
type Plan struct {
	Remove []Entry
	User   []string
	System []string
}

// BuildPlan decides what to remove without performing any I/O.
func BuildPlan(entries []Entry, filter Filter) Plan {
	var plan Plan
	var target string
	if filter.Entry != "" {
		target = Normalize(Expand(filter.Entry))
	}

	for _, entry := range entries {
		if !matchStatus(entry.Status, filter.Only) {
			plan.add(entry)
			continue
		}
		if target != "" && Normalize(entry.Expanded) != target {
			plan.add(entry)
			continue
		}
		plan.Remove = append(plan.Remove, entry)
	}

	return plan
}

func (p *Plan) add(entry Entry) {
	if entry.Scope == ScopeSystem {
		p.System = append(p.System, entry.Path)
		return
	}
	p.User = append(p.User, entry.Path)
}

func matchStatus(status, only Status) bool {
	if only != "" {
		return status == only
	}
	return status == StatusDead || status == StatusDup
}

// RemovedIn returns how many entries are scheduled for removal in a scope.
func (p Plan) RemovedIn(scope Scope) int {
	count := 0
	for _, entry := range p.Remove {
		if entry.Scope == scope {
			count++
		}
	}
	return count
}
