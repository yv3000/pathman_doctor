package main

import (
	"fmt"
	"os"

	"github.com/yv3000/pathman_doctor/cmd"
)

const version = "1.1.0"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "doctor":
		cmd.Doctor()
	case "fix":
		cmd.Fix()
	case "uninstall":
		cmd.Uninstall()
	case "--version", "-v":
		fmt.Printf("pathman v%s\n", version)
	case "--help", "-h":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`pathman — Windows PATH manager

Usage:
  pathman doctor                        Scan PATH and show all issues
  pathman fix                           Auto-remove all dead and duplicate entries
  pathman fix --only dead               Remove only dead entries
  pathman fix --only dup                Remove only duplicate entries
  pathman fix --entry "C:\some\path"    Fix one specific path entry
  pathman --version                     Show version
  pathman --help                        Show this help`)
}
