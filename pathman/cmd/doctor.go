package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/yv3000/pathman/internal"
)

func Doctor() {
	fmt.Println("pathman doctor — scanning PATH entries...\n")

	systemEntries, err := internal.ReadPATH("System")
	if err != nil {
		color.Yellow("⚠️  Could not read System PATH (try running as Administrator)")
	}

	userEntries, err := internal.ReadPATH("User")
	if err != nil {
		color.Yellow("⚠️  Could not read User PATH")
		userEntries = []string{}
	}

	if len(systemEntries) == 0 && len(userEntries) == 0 {
		color.Yellow("⚠️  PATH is empty")
		return
	}

	entries := internal.AnalyzePATH(userEntries, systemEntries)

	deadCount := 0
	dupCount := 0

	fmt.Printf("%-8s %-8s %s\n", "SCOPE", "STATUS", "PATH")
	fmt.Printf("%-8s %-8s %s\n", "------", "------", "----")

	for _, e := range entries {
		var statusText string
		switch e.Status {
		case "OK":
			statusText = color.GreenString("✅ OK    ")
		case "DEAD":
			statusText = color.RedString("❌ DEAD  ")
			deadCount++
		case "DUP":
			statusText = color.YellowString("⚠️  DUP   ")
			dupCount++
		}

		pathText := e.Path
		if e.Note != "" {
			pathText = fmt.Sprintf("%s   %s", e.Path, e.Note)
		}

		fmt.Printf("%-8s %s %s\n", e.Scope, statusText, pathText)
	}

	fmt.Println()

	if deadCount == 0 && dupCount == 0 {
		color.Green("✅ PATH looks clean. Nothing to fix.")
	} else {
		fmt.Printf("Summary: %d entries | %d dead | %d duplicate | action needed\n", len(entries), deadCount, dupCount)
		fmt.Println("Run `pathman fix` to auto-clean.")
	}
}
