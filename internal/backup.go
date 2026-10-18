package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DefaultBackupDir is where PATH snapshots are written before a change.
func DefaultBackupDir() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "pathman", "backups")
}

// SaveBackup writes the raw registry value to dir and returns the file path.
// The file holds exactly what was stored before the change, so restoring it
// is a plain copy back into PATH.
func SaveBackup(dir string, scope Scope, raw string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("no backup directory configured")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}

	name := fmt.Sprintf("PATH-%s-%s.txt", scope, time.Now().Format("20060102-150405.000"))
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		return "", err
	}

	return path, nil
}
