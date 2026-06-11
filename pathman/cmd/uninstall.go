package cmd

import (
	"fmt"
	"os"
	"os/exec"
)

func Uninstall() {
	fmt.Println("pathman uninstall — removing pathman from your system...")
	fmt.Print("Are you sure you want to uninstall pathman? (y/n): ")

	var input string
	fmt.Scanln(&input)
	if input != "y" {
		fmt.Println("Cancelled.")
		return
	}

	psScript := `irm https://raw.githubusercontent.com/yv3000/pathman/main/uninstall.ps1 | iex`
	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Uninstall failed: %v\n", err)
		fmt.Println("Manual removal: delete %USERPROFILE%\\.pathman\\bin and remove it from your User PATH.")
		os.Exit(1)
	}
}
