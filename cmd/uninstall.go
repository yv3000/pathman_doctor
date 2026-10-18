package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/yv3000/pathman_doctor/internal"
)

func defaultExecutable() (string, error) {
	return os.Executable()
}

func defaultRemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (a App) runUninstall(args []string) int {
	if hasHelpFlag(args) {
		a.printUninstallHelp()
		return ExitOK
	}
	if len(args) > 0 {
		a.errorf("unexpected argument %q", args[0])
		a.printUninstallHelp()
		return ExitUsage
	}

	a.printf("pathman uninstall — removing pathman from your system...")

	executable, err := a.Executable()
	if err != nil {
		a.errorf("Could not locate the running pathman binary: %v", err)
		return ExitFailure
	}

	installDir := filepath.Dir(executable)
	if !isInstallDir(installDir) {
		a.errorf("pathman is not running from %%USERPROFILE%%\\.pathman\\bin, so nothing was removed.")
		a.printf("Delete pathman manually and remove its folder from your User PATH.")
		return ExitFailure
	}

	if !a.confirm("Are you sure you want to uninstall pathman?") {
		a.printf("Cancelled.")
		return ExitOK
	}

	entries, err := a.Store.ReadPATH(internal.ScopeUser)
	if err != nil {
		a.errorf("Could not read User PATH: %v", err)
		return ExitFailure
	}

	target := internal.Normalize(installDir)
	kept := make([]string, 0, len(entries))
	removed := 0
	for _, entry := range entries {
		if internal.Normalize(internal.Expand(entry)) == target {
			removed++
			continue
		}
		kept = append(kept, entry)
	}

	if removed > 0 {
		raw, err := a.Store.ReadRaw(internal.ScopeUser)
		if err != nil {
			a.errorf("Could not read User PATH: %v", err)
			return ExitFailure
		}
		if _, err := internal.SaveBackup(a.BackupDir, internal.ScopeUser, raw); err != nil {
			a.errorf("Could not back up User PATH, no changes made: %v", err)
			return ExitFailure
		}
		if err := a.Store.WritePATH(internal.ScopeUser, kept); err != nil {
			a.errorf("Could not remove pathman from User PATH: %v", err)
			return ExitFailure
		}
		a.Store.Broadcast()
		a.printf("Removed %s from User PATH.", installDir)
	}

	if err := a.RemoveAll(installDir); err != nil {
		a.warnf("Could not delete %s: %v", installDir, err)
		a.printf("Close the running pathman process and delete the folder manually.")
		return ExitOK
	}

	a.successf("pathman uninstalled.")
	a.printf("Restart your terminal to finish removal.")
	return ExitOK
}

func (a App) printUninstallHelp() {
	a.printf("Usage: pathman uninstall\n\nRemove pathman's files and its User PATH entry.")
}

// isInstallDir guards against deleting a directory pathman does not own.
func isInstallDir(dir string) bool {
	return strings.EqualFold(filepath.Base(dir), "bin") &&
		strings.EqualFold(filepath.Base(filepath.Dir(dir)), ".pathman")
}
