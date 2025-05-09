package query

import (
	"testing"
)

func TestGetFieldAlias(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Standard aliases
		{name: "name alias", input: "name", expected: "metadata.name"},
		{name: "namespace alias", input: "namespace", expected: "metadata.namespace"},
		{name: "labels alias", input: "labels", expected: "metadata.labels"},
		{name: "phase alias", input: "phase", expected: "status.phase"},

		// Case insensitivity
		{name: "uppercase alias", input: "NAME", expected: "metadata.name"},
		{name: "mixed case alias", input: "NaMeSpAcE", expected: "metadata.namespace"},

		// Whitespace handling
		{name: "trailing space", input: "name ", expected: "metadata.name"},
		{name: "leading space", input: " phase", expected: "status.phase"},
		{name: "both spaces", input: " created ", expected: "metadata.creationTimestamp"},

		// Dot trimming
		{name: "trailing dot", input: "name.", expected: "metadata.name"},
		{name: "leading dot", input: ".labels", expected: "metadata.labels"},
		{name: "both dots", input: ".annotations.", expected: "metadata.annotations"},

		// Combined cases
		{name: "complex case", input: " .NaMe. ", expected: "metadata.name"},

		// Non-aliased fields
		{name: "non-aliased field", input: "spec.replicas", expected: "spec.replicas"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDefaultFieldAlias(tt.input)
			if result != tt.expected {
				t.Errorf("GetFieldAlias(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
