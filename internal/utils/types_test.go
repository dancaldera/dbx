package utils

import "testing"

func TestInferFieldType(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"null value", "NULL", "NULL"},
		{"empty string", "", "Text"},
		{"boolean true", "true", "Bool"},
		{"boolean TRUE", "TRUE", "Bool"},
		{"boolean false", "false", "Bool"},
		{"boolean FALSE", "FALSE", "Bool"},
		{"integer", "42", "Int"},
		{"negative integer", "-123", "Int"},
		{"float", "3.14", "Float"},
		{"negative float", "-2.5", "Float"},
		{"json object", "{\"key\": \"value\"}", "JSON"},
		{"json array", "[1, 2, 3]", "JSON"},
		{"datetime RFC3339", "2023-01-15T10:30:00Z", "DateTime"},
		{"datetime simple", "2023-01-15", "DateTime"},
		{"datetime with time", "2023-01-15 10:30:00", "DateTime"},
		{"plain text", "hello world", "Text"},
		{"text with numbers", "hello123", "Text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InferFieldType(tt.value); got != tt.want {
				t.Errorf("InferFieldType(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestLooksLikeDateTime(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"empty string", "", false},
		{"RFC3339", "2023-01-15T10:30:00Z", true},
		{"RFC3339Nano", "2023-01-15T10:30:00.123456789Z", true},
		{"simple date", "2023-01-15", true},
		{"datetime with seconds", "2023-01-15 10:30:00", true},
		{"datetime with milliseconds", "2023-01-15 10:30:00.000", true},
		{"plain text", "hello", false},
		{"number", "123", false},
		{"very long string", "this is a very long string that exceeds the 64 character limit for datetime detection so it should return false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LooksLikeDateTime(tt.value); got != tt.want {
				t.Errorf("LooksLikeDateTime(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestSanitizeValueForDisplay(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"single line", "hello world", "hello world"},
		{"multiple spaces", "hello    world", "hello world"},
		{"newlines", "hello\nworld", "hello world"},
		{"tabs", "hello\tworld", "hello world"},
		{"mixed whitespace", "hello\n\t  world", "hello world"},
		{"leading/trailing spaces", "  hello world  ", "hello world"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeValueForDisplay(tt.value); got != tt.want {
				t.Errorf("SanitizeValueForDisplay(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestTruncateWithEllipsis(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		budget   int
		ellipsis string
		wantMin  int // minimum expected length
	}{
		{"short string", "hello", 10, "...", 5},
		{"exact length", "hello", 5, "...", 5},
		{"needs truncation", "hello world", 8, "...", 5},
		{"empty string", "", 5, "...", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateWithEllipsis(tt.value, tt.budget, tt.ellipsis)
			if len(got) < tt.wantMin {
				t.Errorf("TruncateWithEllipsis(%q, %d, %q) length = %d, want >= %d", tt.value, tt.budget, tt.ellipsis, len(got), tt.wantMin)
			}
			// For short strings, should return original
			if len(tt.value) <= tt.budget && got != tt.value {
				t.Errorf("TruncateWithEllipsis(%q, %d, %q) = %q, want original string", tt.value, tt.budget, tt.ellipsis, got)
			}
		})
	}
}
