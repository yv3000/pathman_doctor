package main

import "testing"

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain version", in: "1.1.0", want: "1.1.0"},
		{name: "tag version", in: "v1.2.0", want: "1.2.0"},
		{name: "dev build", in: "dev", want: "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeVersion(tt.in); got != tt.want {
				t.Fatalf("normalizeVersion(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
