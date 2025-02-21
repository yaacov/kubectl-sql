package eval

import (
	"testing"
	"time"
)

func TestStringValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"integer", "123", int64(123)},
		{"float", "123.45", float64(123.45)},
		{"boolean true", "true", true},
		{"boolean True", "True", true},
		{"boolean false", "false", false},
		{"date RFC3339", "2020-01-02T15:04:05Z", func() time.Time { t, _ := time.Parse(time.RFC3339, "2020-01-02T15:04:05Z"); return t }()},
		{"date short", "2020-01-02", func() time.Time { t, _ := time.Parse("2006-01-02", "2020-01-02"); return t }()},
		{"string", "hello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringValue(tt.input)
			if got != tt.expected {
				t.Errorf("stringValue(%s) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseSINumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"kilobyte", "1K", float64(1000)},
		{"kibibyte", "1Ki", float64(1024)},
		{"megabyte", "1M", float64(1000000)},
		{"mebibyte", "1Mi", float64(1048576)},
		{"gigabyte", "1G", float64(1000000000)},
		{"terabyte", "1T", float64(1000000000000)},
		{"petabyte", "1P", float64(1000000000000000)},
		{"invalid", "1X", nil},
		{"not SI", "123", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSINumber(tt.input)
			if got != tt.expected {
				t.Errorf("parseSINumber(%s) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}
