package utils

import (
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// InferFieldType detects the data type of a field value
func InferFieldType(v string) string {
	if v == "NULL" {
		return "NULL"
	}
	if v == "" {
		return "Text"
	}
	// Boolean
	if v == "true" || v == "false" || v == "TRUE" || v == "FALSE" {
		return "Bool"
	}
	// Numeric
	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return "Int"
	}
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return "Float"
	}
	// JSON
	s := strings.TrimSpace(v)
	if (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) || (strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]")) {
		return "JSON"
	}
	// DateTime: try to parse with common layouts instead of loose punctuation checks
	if LooksLikeDateTime(strings.TrimSpace(v)) {
		return "DateTime"
	}
	return "Text"
}

// LooksLikeDateTime attempts to detect datetime format
func LooksLikeDateTime(s string) bool {
	if s == "" {
		return false
	}
	// Avoid obviously long textual content
	if len(s) > 64 {
		return false
	}
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RubyDate,
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05 MST",
		"2006-01-02 15:04:05.000 -0700 MST",
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if _, err := time.Parse(layout, s); err == nil {
			return true
		}
	}
	return false
}

// SanitizeValueForDisplay cleans values for single-line UI display
func SanitizeValueForDisplay(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

// TruncateWithEllipsis truncates a string to fit within budget with ellipsis
func TruncateWithEllipsis(value string, budget int, ellipsis string) string {
	return ansi.Truncate(value, budget, ellipsis)
}

// FormatFieldValue formats field values for display, with special handling for JSON
func FormatFieldValue(value string) string {
	// Try to format JSON for better readability
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		// Not JSON, return as-is
		return value
	}

	// Pretty-print JSON
	var formatted strings.Builder
	indent := 0
	inString := false
	escaped := false

	for i, char := range value {
		if escaped {
			formatted.WriteRune(char)
			escaped = false
			continue
		}

		if char == '\\' && inString {
			formatted.WriteRune(char)
			escaped = true
			continue
		}

		if char == '"' {
			inString = !inString
			formatted.WriteRune(char)
			continue
		}

		if inString {
			formatted.WriteRune(char)
			continue
		}

		switch char {
		case '{', '[':
			formatted.WriteRune(char)
			formatted.WriteRune('\n')
			indent++
			for j := 0; j < indent*2; j++ {
				formatted.WriteRune(' ')
			}
		case '}', ']':
			if i > 0 && value[i-1] != '\n' {
				formatted.WriteRune('\n')
			}
			indent--
			for j := 0; j < indent*2; j++ {
				formatted.WriteRune(' ')
			}
			formatted.WriteRune(char)
			if i < len(value)-1 {
				formatted.WriteRune('\n')
				for j := 0; j < indent*2; j++ {
					formatted.WriteRune(' ')
				}
			}
		case ',':
			formatted.WriteRune(char)
			formatted.WriteRune('\n')
			for j := 0; j < indent*2; j++ {
				formatted.WriteRune(' ')
			}
		case ':':
			formatted.WriteRune(char)
			formatted.WriteRune(' ')
		default:
			if char != ' ' || formatted.Len() == 0 || formatted.String()[formatted.Len()-1] != ' ' {
				formatted.WriteRune(char)
			}
		}
	}

	return formatted.String()
}
