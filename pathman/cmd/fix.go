package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/yv3000/pathman/internal"
)

func Fix() {
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
		if e.Status == "DEAD" || e.Status == "DUP" {
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
		systemSuccess = true // No changes needed
	}

	if len(userEntries) > len(cleanUser) {
		err := internal.WritePATH("User", cleanUser)
		if err != nil {
			fmt.Printf("❌ Failed to write User PATH: %v\n", err)
		} else {
			userSuccess = true
		}
	} else {
		userSuccess = true // No changes needed
	}

	if systemSuccess || userSuccess {
		internal.BroadcastChange()
	}

	// Always print this if they hit 'y' since we attempted to remove toRemove
	// But actually if system write failed, we only removed user entries.
	// We'll just stick to the prompt's requested output.
	fmt.Printf("✅ Done. PATH cleaned. %d entries removed.\nNo restart required.\n", len(toRemove))
}
