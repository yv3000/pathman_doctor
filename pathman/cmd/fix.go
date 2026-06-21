package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/yv3000/pathman_doctor/internal"
)

func Fix() {
	// Parse optional flags: --entry <path>, --only dead|dup
	var entryFilter string
	var onlyFilter string

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--entry":
			if i+1 < len(args) {
				entryFilter = strings.TrimRight(args[i+1], `\/`)
				i++
			}
		case "--only":
			if i+1 < len(args) {
				onlyFilter = strings.ToUpper(args[i+1])
				i++
			}
		}
	}

	systemEntries, err := internal.ReadPATH("System")
	if err != nil {
		// proceed with what we have
	}

	userEntries, err := internal.ReadPATH("User")
	if err != nil {
		userEntries = []string{}
	}

	entries := internal.AnalyzePATH(userEntries, systemEntries)

	var toRemove []internal.Entry
	var cleanSystem []string
	var cleanUser []string

	for _, e := range entries {
		isIssue := e.Status == "DEAD" || e.Status == "DUP"

		if isIssue && entryFilter != "" {
			normalized := strings.TrimRight(e.Path, `\/`)
			if !strings.EqualFold(normalized, entryFilter) {
				isIssue = false
			}
		}
		if isIssue && onlyFilter != "" {
			if !strings.EqualFold(e.Status, onlyFilter) {
				isIssue = false
			}
		}

		if isIssue {
			toRemove = append(toRemove, e)
		} else {
			if e.Scope == "System" {
				cleanSystem = append(cleanSystem, e.Path)
			} else {
				cleanUser = append(cleanUser, e.Path)
			}
		}
	}

	if len(toRemove) == 0 {
		color.Green("✅ PATH looks clean. Nothing to fix.")
		return
	}

	fmt.Printf("pathman fix — found %d issues\n\n", len(toRemove))

	for _, e := range toRemove {
		if e.Status == "DEAD" {
			fmt.Printf("  %s %s   %s\n", color.RedString("❌ Removing DEAD: "), e.Path, e.Note)
		} else if e.Status == "DUP" {
			fmt.Printf("  %s  %s   (keeping first occurrence)\n", color.YellowString("⚠️  Removing DUP: "), e.Path)
		}
	}

	fmt.Print("\nApply these changes? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input != "y" {
		fmt.Println("Cancelled. No changes made.")
		return
	}

	systemSuccess := false
	userSuccess := false

	if len(systemEntries) > len(cleanSystem) {
		err := internal.WritePATH("System", cleanSystem)
		if err != nil {
			color.Yellow("⚠️  System PATH requires admin rights. Run as Administrator to fix System entries.")
		} else {
			systemSuccess = true
		}
	} else {
		systemSuccess = true
	}

	if len(userEntries) > len(cleanUser) {
		err := internal.WritePATH("User", cleanUser)
		if err != nil {
			fmt.Printf("❌ Failed to write User PATH: %v\n", err)
		} else {
			userSuccess = true
		}
	} else {
		userSuccess = true
	}

	if systemSuccess || userSuccess {
		internal.BroadcastChange()
		fmt.Printf("✅ Done. PATH cleaned. %d entries removed.\nNo restart required.\n", len(toRemove))
	} else {
		fmt.Println("❌ No changes applied. Check the errors above.")
	}
}
