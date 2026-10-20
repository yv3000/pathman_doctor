package internal

import (
	"fmt"
	"strings"
)

// ValidateEntry normalizes a PATH value supplied on the command line and
// rejects values that cannot be stored as one PATH entry. Keeping the rule in
// one place means every command rejects the same input for the same reason.
func ValidateEntry(raw string) (string, error) {
	entry := strings.TrimSpace(raw)
	if entry == "" {
		return "", fmt.Errorf("PATH entry is empty")
	}
	if strings.ContainsRune(entry, ';') {
		return "", fmt.Errorf("PATH entry %q contains ';', which separates entries", entry)
	}
	if strings.ContainsRune(entry, 0) {
		return "", fmt.Errorf("PATH entry %q contains a NUL byte", entry)
	}
	return entry, nil
}
