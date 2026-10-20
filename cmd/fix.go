package cmd

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/yv3000/pathman_doctor/internal"
)

type fixOptions struct {
	Entry string
	Only  internal.Status
}

func parseFixArgs(args []string) (fixOptions, error) {
	flags := flag.NewFlagSet("fix", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	entry := flags.String("entry", "", `only fix this PATH entry`)
	only := flags.String("only", "", `remove only "dead" or "dup" entries`)

	if err := flags.Parse(args); err != nil {
		return fixOptions{}, err
	}
	if flags.NArg() > 0 {
		return fixOptions{}, fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}

	entrySet := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "entry" {
			entrySet = true
		}
	})

	opts := fixOptions{}
	if entrySet {
		validated, err := internal.ValidateEntry(*entry)
		if err != nil {
			return fixOptions{}, err
		}
		opts.Entry = validated
	}

	switch strings.ToLower(strings.TrimSpace(*only)) {
	case "":
	case "dead":
		opts.Only = internal.StatusDead
	case "dup":
		opts.Only = internal.StatusDup
	default:
		return fixOptions{}, fmt.Errorf("invalid --only value %q (want dead or dup)", *only)
	}

	return opts, nil
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

	opts, err := parseFixArgs(args)
	if err != nil {
		a.errorf("%v", err)
		a.printFixHelp()
		return ExitUsage
	}

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

	backups, err := a.saveBackups(plan)
	if err != nil {
		a.errorf("Could not back up PATH, no changes made: %v", err)
		return ExitFailure
	}
	if len(backups) > 0 {
		a.printf("Backup: %s", strings.Join(backups, ", "))
	}

	removed := 0
	failed := make([]string, 0, 2)

	if plan.RemovedIn(internal.ScopeSystem) > 0 {
		if err := a.Store.WritePATH(internal.ScopeSystem, plan.System); err != nil {
			a.warnf("System PATH requires admin rights. Run as Administrator to fix System entries.")
			failed = append(failed, "System")
		} else {
			removed += plan.RemovedIn(internal.ScopeSystem)
		}
	}

	if plan.RemovedIn(internal.ScopeUser) > 0 {
		if err := a.Store.WritePATH(internal.ScopeUser, plan.User); err != nil {
			a.errorf("Failed to write User PATH: %v", err)
			failed = append(failed, "User")
		} else {
			removed += plan.RemovedIn(internal.ScopeUser)
		}
	}

	if len(failed) == 0 {
		a.Store.Broadcast()
		a.successf("PATH cleaned. %d entries removed.", removed)
		a.printf("No restart required.")
		return ExitOK
	}

	if removed > 0 {
		a.Store.Broadcast()
		a.warnf("Removed %d entries, but the %s PATH was left unchanged.", removed, strings.Join(failed, " and "))
		return ExitFailure
	}

	a.errorf("No changes applied. Check the errors above.")
	return ExitFailure
}

func (a App) saveBackups(plan internal.Plan) ([]string, error) {
	var paths []string
	for _, scope := range []internal.Scope{internal.ScopeSystem, internal.ScopeUser} {
		if plan.RemovedIn(scope) == 0 {
			continue
		}
		raw, err := a.Store.ReadRaw(scope)
		if err != nil {
			return nil, fmt.Errorf("read %s PATH: %w", scope, err)
		}
		path, err := internal.SaveBackup(a.BackupDir, scope, raw)
		if err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}
