package internal

import "testing"

func TestValidateEntry(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "plain path", in: `C:\Tools\bin`, want: `C:\Tools\bin`},
		{name: "trims surrounding space", in: "  C:\\Tools\\bin\t", want: `C:\Tools\bin`},
		{name: "percent variables are allowed", in: `%USERPROFILE%\bin`, want: `%USERPROFILE%\bin`},
		{name: "empty", in: "", wantErr: true},
		{name: "whitespace only", in: "   ", wantErr: true},
		{name: "entry separator", in: `C:\one;C:\two`, wantErr: true},
		{name: "nul byte", in: "C:\\one\x00", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateEntry(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ValidateEntry(%q) = %q, want an error", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateEntry(%q) = %v", tt.in, err)
			}
			if got != tt.want {
				t.Fatalf("ValidateEntry(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestValidateOnlyFlag(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    Status
		wantErr bool
	}{
		{name: "unset keeps the default filter", in: "", want: ""},
		{name: "whitespace keeps the default filter", in: "   ", want: ""},
		{name: "dead", in: "dead", want: StatusDead},
		{name: "duplicate", in: "dup", want: StatusDup},
		{name: "case and space insensitive", in: "  DUP ", want: StatusDup},
		{name: "unknown value", in: "bogus", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateOnlyFlag(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ValidateOnlyFlag(%q) = %q, want an error", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateOnlyFlag(%q) = %v", tt.in, err)
			}
			if got != tt.want {
				t.Fatalf("ValidateOnlyFlag(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
