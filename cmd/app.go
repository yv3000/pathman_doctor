package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/yv3000/pathman_doctor/internal"
)

// Exit codes shared by every subcommand.
const (
	ExitOK      = 0
	ExitFailure = 1
	ExitUsage   = 2
)

var (
	green  = color.New(color.FgGreen)
	red    = color.New(color.FgRed)
	yellow = color.New(color.FgYellow)
)

// App holds everything a command needs. Tests replace Store and the streams.
type App struct {
	Store   internal.Store
	In      io.Reader
	Out     io.Writer
	Version string

	// BackupDir is where PATH is snapshotted before a write.
	BackupDir string
	// Executable, RemoveAll and ScheduleRemoval are seams for uninstall.
	Executable      func() (string, error)
	RemoveAll       func(string) error
	ScheduleRemoval func(string) error
}

// NewApp builds the production application.
func NewApp(store internal.Store, in io.Reader, out io.Writer, version string) App {
	return App{
		Store:           store,
		In:              in,
		Out:             out,
		Version:         version,
		BackupDir:       internal.DefaultBackupDir(),
		Executable:      defaultExecutable,
		RemoveAll:       defaultRemoveAll,
		ScheduleRemoval: defaultScheduleRemoval,
	}
}

// Run dispatches a command line and returns the process exit code.
func (a App) Run(args []string) int {
	if len(args) == 0 {
		a.printHelp()
		return ExitOK
	}

	switch args[0] {
	case "doctor":
		return a.runDoctor(args[1:])
	case "fix":
		return a.runFix(args[1:])
	case "uninstall":
		return a.runUninstall(args[1:])
	case "--version", "-v":
		a.printf("pathman v%s", a.Version)
		return ExitOK
	case "--help", "-h":
		a.printHelp()
		return ExitOK
	default:
		a.errorf("Unknown command: %s", args[0])
		a.printHelp()
		return ExitUsage
	}
}

func (a App) printHelp() {
	a.printf(`pathman — Windows PATH manager

Usage:
  pathman doctor                        Scan PATH and show all issues
  pathman fix                           Auto-remove all dead and duplicate entries
  pathman fix --only dead               Remove only dead entries
  pathman fix --only dup                Remove only duplicate entries
  pathman fix --entry "C:\some\path"    Fix one specific path entry
  pathman uninstall                     Remove pathman from this machine
  pathman --version                     Show version
  pathman --help                        Show this help

Exit codes: 0 success, 1 incomplete or failed, 2 invalid usage.`)
}

func (a App) printf(format string, args ...any) {
	_, _ = fmt.Fprintf(a.Out, format+"\n", args...)
}

func (a App) warnf(format string, args ...any) {
	_, _ = yellow.Fprintf(a.Out, "⚠️  "+format+"\n", args...)
}

func (a App) errorf(format string, args ...any) {
	_, _ = red.Fprintf(a.Out, "❌ "+format+"\n", args...)
}

func (a App) confirm(prompt string) bool {
	_, _ = fmt.Fprintf(a.Out, "%s (y/n): ", prompt)
	line, err := bufio.NewReader(a.In).ReadString('\n')
	if err != nil && line == "" {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(line), "y")
}

func (a App) successf(format string, args ...any) {
	_, _ = green.Fprintf(a.Out, "✅ "+format+"\n", args...)
}

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}
