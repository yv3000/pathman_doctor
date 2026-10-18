package cmd

import (
	"strings"

	"github.com/yv3000/pathman_doctor/internal"
)

// invalidStatus is used while parsing an unknown --only value so that a typo
// never turns into "remove everything". Flag validation rejects it later.
const invalidStatus internal.Status = "INVALID"

type fixOptions struct {
	Entry string
	Only  internal.Status
}

func parseFixArgs(args []string) fixOptions {
	var opts fixOptions
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--entry":
			if i+1 < len(args) {
				opts.Entry = strings.TrimRight(strings.TrimSpace(args[i+1]), `\/`)
				i++
			}
		case "--only":
			if i+1 < len(args) {
				opts.Only = parseOnly(args[i+1])
				i++
			}
		}
	}
	return opts
}

func parseOnly(value string) internal.Status {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "dead":
		return internal.StatusDead
	case "dup":
		return internal.StatusDup
	default:
		return invalidStatus
	}
}

func (a App) printFixHelp() {
	a.printf(`Usage: pathman fix [--only dead|dup] [--entry "C:\path"]

Removes dead and duplicate PATH entries after showing what will change and
asking for confirmation.`)
}

func (a App) runFix(args []string) int {
	if hasHelpFlag(args) {
		a.printFixHelp()
		return ExitOK
	}

	opts := parseFixArgs(args)

	systemEntries, sysErr := a.Store.ReadPATH(internal.ScopeSystem)
	if sysErr != nil {
		a.warnf("Could not read System PATH (try running as Administrator)")
	}

	userEntries, userErr := a.Store.ReadPATH(internal.ScopeUser)
	if userErr != nil {
		a.warnf("Could not read User PATH")
		userEntries = nil
	}

	if sysErr != nil && userErr != nil {
		a.errorf("Could not read either PATH scope; nothing to fix.")
		return ExitFailure
	}

	entries := internal.AnalyzePATH(userEntries, systemEntries)
	plan := internal.BuildPlan(entries, internal.Filter{Only: opts.Only, Entry: opts.Entry})

	if len(plan.Remove) == 0 {
		a.successf("PATH looks clean. Nothing to fix.")
		return ExitOK
	}

	a.printf("pathman fix — found %d issues\n", len(plan.Remove))

	for _, entry := range plan.Remove {
		switch entry.Status {
		case internal.StatusDead:
			a.printf("  %s %s   %s", red.Sprint("❌ Removing DEAD: "), entry.Path, entry.Note)
		case internal.StatusDup:
			a.printf("  %s  %s   (keeping first occurrence)", yellow.Sprint("⚠️  Removing DUP: "), entry.Path)
		}
	}

	if !a.confirm("\nApply these changes?") {
		a.printf("Cancelled. No changes made.")
		return ExitOK
	}

	systemOK := true
	userOK := true

	if plan.RemovedIn(internal.ScopeSystem) > 0 {
		if err := a.Store.WritePATH(internal.ScopeSystem, plan.System); err != nil {
			a.warnf("System PATH requires admin rights. Run as Administrator to fix System entries.")
			systemOK = false
		}
	}

	if plan.RemovedIn(internal.ScopeUser) > 0 {
		if err := a.Store.WritePATH(internal.ScopeUser, plan.User); err != nil {
			a.errorf("Failed to write User PATH: %v", err)
			userOK = false
		}
	}

	if systemOK || userOK {
		a.Store.Broadcast()
		a.printf("✅ Done. PATH cleaned. %d entries removed.", len(plan.Remove))
		a.printf("No restart required.")
		return ExitOK
	}

	a.errorf("No changes applied. Check the errors above.")
	return ExitFailure
}
