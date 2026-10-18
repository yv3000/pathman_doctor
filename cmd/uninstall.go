package cmd

import (
	"os/exec"
)

func (a App) runUninstall(args []string) int {
	if hasHelpFlag(args) {
		a.printf("Usage: pathman uninstall\n\nRemove pathman from this machine.")
		return ExitOK
	}

	a.printf("pathman uninstall — removing pathman from your system...")

	if !a.confirm("Are you sure you want to uninstall pathman?") {
		a.printf("Cancelled.")
		return ExitOK
	}

	psScript := `irm https://raw.githubusercontent.com/yv3000/pathman_doctor/main/pathman/uninstall.ps1 | iex`
	command := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psScript)
	command.Stdout = a.Out
	command.Stderr = a.Out
	if err := command.Run(); err != nil {
		a.errorf("Uninstall failed: %v", err)
		a.printf("Manual removal: delete %%USERPROFILE%%\\.pathman\\bin and remove it from your User PATH.")
		return ExitFailure
	}

	return ExitOK
}
