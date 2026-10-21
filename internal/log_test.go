package internal

import (
	"bytes"
	"errors"
	"testing"
)

func TestLogfTagsEveryLevel(t *testing.T) {
	tests := []struct {
		name   string
		level  Level
		format string
		args   []any
		want   string
	}{
		{name: "info", level: LevelInfo, format: "reading %s", args: []any{"PATH"}, want: "[INFO] reading PATH\n"},
		{name: "warn", level: LevelWarn, format: "System PATH unreadable", want: "[WARN] System PATH unreadable\n"},
		{name: "error", level: LevelError, format: "write failed: %v", args: []any{errors.New("access denied")}, want: "[ERROR] write failed: access denied\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			Logf(&buf, tt.level, tt.format, tt.args...)
			if got := buf.String(); got != tt.want {
				t.Fatalf("Logf() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLogfWithoutWriterDoesNothing(t *testing.T) {
	Logf(nil, LevelError, "discarded %d", 1)
}
