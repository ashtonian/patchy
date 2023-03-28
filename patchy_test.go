package patchy

import (
	"testing"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"TitleCaseString", "title_case_string"},
		// {"HTTPStatusCode", "http_status_code"},
		{"", ""},
		{"a", "a"},
		{"A", "a"},
		{"_", "_"},
	}

	for _, tt := range tests {
		got := ToSnakeCase(tt.input)
		if got != tt.want {
			t.Errorf("toSnakeCase(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}
