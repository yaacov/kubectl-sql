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

		// Whitespace handling
		{name: "trailing space", input: "name ", expected: "metadata.name"},
		{name: "leading space", input: " phase", expected: "status.phase"},
		{name: "both spaces", input: " created ", expected: "metadata.creationTimestamp"},

		// Dot trimming
		{name: "trailing dot", input: "name.", expected: "metadata.name"},
		{name: "leading dot", input: ".labels", expected: "metadata.labels"},
		{name: "both dots", input: ".annotations.", expected: "metadata.annotations"},

		// Non-aliased fields
		{name: "non-aliased field", input: "spec.replicas", expected: "spec.replicas"},

		// Array/Map indexing
		{name: "labels with index", input: "labels[app]", expected: "metadata.labels[app]"},
		{name: "labels with numeric index", input: "labels[0]", expected: "metadata.labels[0]"},
		{name: "labels with index and spaces", input: "labels [app]", expected: "metadata.labels[app]"},
		{name: "non-aliased field with index", input: "spec.containers[0].name", expected: "spec.containers[0].name"},

		// Function expressions
		{name: "len function", input: "len(status.conditions[*])", expected: "len(status.conditions[*])"},
		{name: "max function", input: "max(spec.replicas)", expected: "max(spec.replicas)"},
		{name: "min function with alias", input: "min(metadata.creationTimestamp)", expected: "min(metadata.creationTimestamp)"},

		// Function expressions with field aliases inside
		{name: "len with field alias", input: "len(conditions[*])", expected: "len(status.conditions[*])"},
		{name: "max with field alias", input: "max(replicas)", expected: "max(spec.replicas)"},
		{name: "function with complex alias", input: "min(created)", expected: "min(metadata.creationTimestamp)"},
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
