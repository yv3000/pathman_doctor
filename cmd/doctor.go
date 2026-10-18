package cmd

import (
	"fmt"

	"github.com/yv3000/pathman_doctor/internal"
)

func (a App) runDoctor(args []string) int {
	if hasHelpFlag(args) {
		a.printf("Usage: pathman doctor\n\nScan the System and User PATH and report dead and duplicate entries.")
		return ExitOK
	}

	a.printf("pathman doctor — scanning PATH entries...")
	a.printf("")

	systemEntries, sysErr := a.Store.ReadPATH(internal.ScopeSystem)
	if sysErr != nil {
		a.warnf("Could not read System PATH (try running as Administrator)")
	}

	userEntries, userErr := a.Store.ReadPATH(internal.ScopeUser)
	if userErr != nil {
		a.warnf("Could not read User PATH")
		userEntries = nil
	}

	if len(systemEntries) == 0 && len(userEntries) == 0 {
		a.warnf("PATH is empty")
		return ExitOK
	}

	entries := internal.AnalyzePATH(userEntries, systemEntries)

	deadCount := 0
	dupCount := 0

	a.printf("%-8s %-8s %s", "SCOPE", "STATUS", "PATH")
	a.printf("%-8s %-8s %s", "------", "------", "----")

	for _, entry := range entries {
		var statusText string
		switch entry.Status {
		case internal.StatusOK:
			statusText = green.Sprint("✅ OK    ")
		case internal.StatusDead:
			statusText = red.Sprint("❌ DEAD  ")
			deadCount++
		case internal.StatusDup:
			statusText = yellow.Sprint("⚠️  DUP   ")
			dupCount++
		}

		pathText := entry.Path
		if entry.Note != "" {
			pathText = fmt.Sprintf("%s   %s", entry.Path, entry.Note)
		}

		a.printf("%-8s %s %s", entry.Scope, statusText, pathText)
	}

	a.printf("")

	if deadCount == 0 && dupCount == 0 {
		a.successf("PATH looks clean. Nothing to fix.")
		return ExitOK
	}

	a.printf("Summary: %d entries | %d dead | %d duplicate | action needed", len(entries), deadCount, dupCount)
	a.printf("Run `pathman fix` to auto-clean.")
	return ExitOK
}
