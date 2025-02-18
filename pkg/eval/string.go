package eval

import (
	"math"
	"strconv"
	"time"
)

// stringValue parses a string to appropriate type (number, boolean, date, or string)
func stringValue(str string) interface{} {
	if v := parseSINumber(str); v != nil {
		return v
	}
	if v := parseBoolean(str); v != nil {
		return *v
	}
	if v := parseDate(str); v != nil {
		return *v
	}
	return str
}

func parseSINumber(s string) interface{} {
	multiplier := 0.0
	base := 1000.0

	// Check for binary prefix
	if len(s) > 1 && s[len(s)-1:] == "i" {
		base = 1024.0
		s = s[:len(s)-1]
	}

	// Check for SI postfix
	if len(s) > 1 {
		postfix := s[len(s)-1:]
		switch postfix {
		case "K":
			multiplier = base
		case "M":
			multiplier = math.Pow(base, 2)
		case "G":
			multiplier = math.Pow(base, 3)
		case "T":
			multiplier = math.Pow(base, 4)
		case "P":
			multiplier = math.Pow(base, 5)
		}

		if multiplier >= 1.0 {
			s = s[:len(s)-1]
			if i, err := strconv.ParseInt(s, 10, 64); err == nil {
				return float64(i) * multiplier
			}
		}
	}

	return nil
}

func parseBoolean(str string) *bool {
	switch str {
	case "true", "True":
		v := true
		return &v
	case "false", "False":
		v := false
		return &v
	}
	return nil
}

func parseDate(str string) *time.Time {
	if t, err := time.Parse(time.RFC3339, str); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02", str); err == nil {
		return &t
	}
	return nil
}
