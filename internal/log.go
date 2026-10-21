package internal

import (
	"fmt"
	"io"
)

// Level tags a diagnostic line so tooling can tell informational messages from
// warnings and errors.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Logf writes one leveled diagnostic line to w. Diagnostics use a separate sink
// from the colored console output: a bug report can attach them without ANSI
// escapes, and a nil writer is a no-op so callers never need to check.
func Logf(w io.Writer, level Level, format string, args ...any) {
	if w == nil {
		return
	}
	_, _ = fmt.Fprintf(w, "[%s] %s\n", level, fmt.Sprintf(format, args...))
}
