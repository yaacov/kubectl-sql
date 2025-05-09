package query

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// ParseString attempts to parse a string into a more specific type.
// It tries to parse the string as an integer, float, boolean, date, or ISO number.
// If none of these parsings succeed, it returns the original string.
func ParseString(str string) (interface{}, error) {
	// Try to parse as int
	if i, err := strconv.Atoi(str); err == nil {
		return i, nil
	}

	// Try to parse as float
	if f, err := strconv.ParseFloat(str, 64); err == nil {
		return f, nil
	}

	// Try to parse as bool
	if b, err := strconv.ParseBool(str); err == nil {
		return b, nil
	}

	// Try to parse as date
	if t, err := ParseDate(str); err == nil {
		return t, nil
	}

	// Try to parse as ISO number
	if n, err := ParseISO(str); err == nil {
		return n, nil
	}

	// If all parsing attempts fail, return the original string
	return str, nil
}

// ParseDate parses a string as a date.
// It tries several common Kubernetes time formats.
func ParseDate(str string) (any, error) {
	// Try common Kubernetes time formats, starting with RFC3339
	timeFormats := []string{
		time.RFC3339,          // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano,      // 2006-01-02T15:04:05.999999999Z07:00
		"2006-01-02T15:04:05", // RFC3339 without timezone
		"2006-01-02",          // Just date
	}

	for _, format := range timeFormats {
		if t, err := time.Parse(format, str); err == nil {
			return t, nil
		}
	}

	// If none of the formats match, return an error
	return nil, fmt.Errorf("could not parse %q as a date", str)
}

// ParseISO parses a string as a number with an ISO unit suffix
// It handles both SI units (powers of 10) and IEC units (powers of 1024)
func ParseISO(str string) (any, error) {
	// Regex to match a number possibly followed by an ISO suffix
	re := regexp.MustCompile(`^(\d+(\.\d+)?)([kKMGTP]i?)?$`)
	matches := re.FindStringSubmatch(str)

	if matches == nil {
		return nil, fmt.Errorf("could not parse %q as an ISO number", str)
	}

	baseNum, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base number in %q: %v", str, err)
	}

	suffix := matches[3]
	if suffix == "" {
		return nil, fmt.Errorf("no ISO suffix found in %q", str)
	}

	switch suffix {
	// SI units (powers of 10)
	case "k", "K":
		return baseNum * 1e3, nil
	case "M":
		return baseNum * 1e6, nil
	case "G":
		return baseNum * 1e9, nil
	case "T":
		return baseNum * 1e12, nil
	case "P":
		return baseNum * 1e15, nil
	// IEC units (powers of 1024)
	case "Ki":
		return baseNum * 1024, nil
	case "Mi":
		return baseNum * 1024 * 1024, nil
	case "Gi":
		return baseNum * 1024 * 1024 * 1024, nil
	case "Ti":
		return baseNum * 1024 * 1024 * 1024 * 1024, nil
	case "Pi":
		return baseNum * 1024 * 1024 * 1024 * 1024 * 1024, nil
	default:
		return nil, fmt.Errorf("unknown ISO suffix %q in %q", suffix, str)
	}
}
