package internal

import (
	"fmt"
	"os"
	"strings"
)

type Entry struct {
	Scope  string // "User" or "System"
	Path   string // the raw path string
	Status string // "OK", "DEAD", or "DUP"
	Note   string // extra info e.g. "(folder missing)" or "(duplicate — appears 2x)"
}

func AnalyzePATH(userEntries []string, systemEntries []string) []Entry {
	var result []Entry

	counts := make(map[string]int)
	for _, p := range append(systemEntries, userEntries...) {
		expanded := os.ExpandEnv(p)
		lowerPath := strings.ToLower(expanded)
		counts[lowerPath]++
	}

	seen := make(map[string]bool)

	processEntries := func(entries []string, scope string) {
		for _, p := range entries {
			expanded := os.ExpandEnv(p)
			lowerPath := strings.ToLower(expanded)

			status := "OK"
			note := ""

			isDup := false
			if seen[lowerPath] {
				isDup = true
			}
			seen[lowerPath] = true

			if isDup {
				status = "DUP"
				note = fmt.Sprintf("(duplicate — appears %dx)", counts[lowerPath])
			} else {
				_, err := os.Stat(expanded)
				if err != nil {
					status = "DEAD"
					note = "(folder missing)"
				}
			}

			result = append(result, Entry{
				Scope:  scope,
				Path:   p,
				Status: status,
				Note:   note,
			})
		}
	}

	processEntries(systemEntries, "System")
	processEntries(userEntries, "User")

	return result
}
