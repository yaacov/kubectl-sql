/*
Copyright 2020 Yaacov Zamir <kobi.zamir@gmail.com>
and other contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Author: 2020 Yaacov Zamir <kobi.zamir@gmail.com>
*/

package query

import (
	"regexp"
	"strings"
)

// fieldAliases is a map of common field aliases to their full paths
var fieldAliases = map[string]string{
	"name":        "metadata.name",
	"namespace":   "metadata.namespace",
	"labels":      "metadata.labels",
	"annotations": "metadata.annotations",
	"created":     "metadata.creationTimestamp",
	"deleted":     "metadata.deletionTimestamp",
	"phase":       "status.phase",
	"replicas":    "spec.replicas",
	"conditions":  "status.conditions",
}

// Function pattern matches expressions like "len(field)"
var functionPattern = regexp.MustCompile(`^(\w+)\((.*)\)$`)

// Array pattern matches expressions like "field[index]"
var arrayPattern = regexp.MustCompile(`^(.+?)(\[.+\])$`)

// GetDefaultFieldAlias returns the full path for a field alias or the original field if no alias exists
func GetDefaultFieldAlias(field string) string {
	// Check for function format (e.g., len(field))
	if matches := functionPattern.FindStringSubmatch(field); matches != nil {
		functionName := matches[1]
		argument := matches[2]

		// Apply field alias resolution to the argument
		resolvedArgument := GetDefaultFieldAlias(argument)

		// Reconstruct the function call
		return functionName + "(" + resolvedArgument + ")"
	}

	// Check for array format (field[...])
	var baseField string
	var arrayIndex string

	if matches := arrayPattern.FindStringSubmatch(field); matches != nil {
		baseField = matches[1]
		arrayIndex = matches[2]
	} else {
		baseField = field
		arrayIndex = ""
	}

	// Create normalized field name: lowercase, trim spaces and dots from ends
	normalizedField := strings.Trim(baseField, " .")

	// Use the mapped field if it exists, otherwise keep the original
	if fullPath, exists := fieldAliases[normalizedField]; exists {
		return fullPath + arrayIndex
	}

	return field
}
