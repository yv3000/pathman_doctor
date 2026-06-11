package main

import (
	"fmt"
	"os"

	"github.com/yv3000/pathman/cmd"
)

const version = "1.0.0"

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
  pathman doctor        Scan PATH and show all issues
  pathman fix           Auto-remove dead and duplicate entries
  pathman --version     Show version
  pathman --help        Show this help`)
}
