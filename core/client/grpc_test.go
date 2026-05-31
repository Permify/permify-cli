package client

import "testing"

func TestFormatBearerToken(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "raw", in: "abc", want: "Bearer abc"},
		{name: "raw trimmed", in: "  abc  ", want: "Bearer abc"},
		{name: "already bearer", in: "Bearer abc", want: "Bearer abc"},
		{name: "already bearer lower", in: "bearer abc", want: "bearer abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatBearerToken(tt.in); got != tt.want {
				t.Fatalf("formatBearerToken(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
