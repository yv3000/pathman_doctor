package internal

import (
	"strings"
	"testing"

	"golang.org/x/sys/windows/registry"
)

func TestGetRegistryKeyAndName(t *testing.T) {
	tests := []struct {
		name      string
		scope     Scope
		wantRoot  registry.Key
		wantPath  string
		wantValue string
	}{
		{
			name:      "user",
			scope:     ScopeUser,
			wantRoot:  registry.CURRENT_USER,
			wantPath:  `Environment`,
			wantValue: `PATH`,
		},
		{
			name:      "system",
			scope:     ScopeSystem,
			wantRoot:  registry.LOCAL_MACHINE,
			wantPath:  `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
			wantValue: `Path`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, path, value := getRegistryKeyAndName(tt.scope)
			if root != tt.wantRoot {
				t.Fatalf("root = %v, want %v", root, tt.wantRoot)
			}
			if path != tt.wantPath {
				t.Fatalf("path = %q, want %q", path, tt.wantPath)
			}
			if value != tt.wantValue {
				t.Fatalf("value name = %q, want %q", value, tt.wantValue)
			}
		})
	}
}

// TestReadPATHUserIsReadable only reads the current user's hive, so it never
// needs admin rights and never modifies PATH.
func TestReadPATHUserIsReadable(t *testing.T) {
	entries, err := ReadPATH(ScopeUser)
	if err != nil {
		t.Fatalf("ReadPATH(User) = %v", err)
	}
	for _, entry := range entries {
		if strings.TrimSpace(entry) == "" {
			t.Fatalf("ReadPATH(User) returned an empty entry: %#v", entries)
		}
	}
}
