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

package eval

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/jsonpath"
)

// ExtractValue extract a value from an item using a key.
func ExtractValue(item unstructured.Unstructured, key string) (interface{}, bool) {
	// Check for reserved words.
	switch key {
	case "name":
		return item.GetName(), true
	case "namespace":
		return item.GetNamespace(), true
	case "created":
		return item.GetCreationTimestamp().Time.UTC(), true
	case "deleted":
		return item.GetDeletionTimestamp().Time.UTC(), true
	}

	// Check for labels and annotations.
	if strings.HasPrefix(key, "labels.") {
		value, ok := item.GetLabels()[key[7:]]
		return handleMetadataValue(value, ok)
	}

	if strings.HasPrefix(key, "annotations.") {
		value, ok := item.GetAnnotations()[key[12:]]
		return handleMetadataValue(value, ok)
	}

	// Use Kubernetes JSONPath implementation
	// Format the key as a proper JSONPath expression if it's not already
	if !strings.HasPrefix(key, "{") {
		key = fmt.Sprintf("{.%s}", key)
	}

	// Check if the path contains a wildcard pattern
	hasWildcard := strings.Contains(key, "[*]") || strings.Contains(key, "..") ||
		strings.Contains(key, "*") || strings.Contains(key, "?")

	j := jsonpath.New("extract-value")
	if err := j.Parse(key); err != nil {
		return nil, true
	}

	buf := &bytes.Buffer{}
	if err := j.Execute(buf, item.Object); err != nil {
		return nil, true
	}

	// If there's no output, the path doesn't exist
	if buf.Len() == 0 {
		return nil, true
	}

	// Parse the result
	var result interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		trimmedStr := strings.TrimSpace(buf.String())

		if hasWildcard {
			// If the path has a wildcard but the result couldn't be unmarshaled as JSON,
			// split by spaces and create an array
			parts := strings.Fields(trimmedStr)
			convertedArray := make([]interface{}, len(parts))
			for i, part := range parts {
				convertedArray[i] = inferValue(part)
			}
			return convertedArray, true
		}

		convertedValue, _ := convertObjectToValue(trimmedStr)
		return convertedValue, true
	}

	// If wildcard is present, ensure we return an array
	if hasWildcard {
		switch v := result.(type) {
		case []interface{}:
			// Already an array, convert each element
			convertedArray := make([]interface{}, len(v))
			for i, item := range v {
				convertedArray[i], _ = convertObjectToValue(item)
			}
			return convertedArray, true
		default:
			// Convert to array with single element
			converted, _ := convertObjectToValue(result)
			return []interface{}{converted}, true
		}
	}

	// If result is a single value array or map with one entry, extract it
	switch v := result.(type) {
	case []interface{}:
		if len(v) == 0 {
			return []interface{}{}, true
		}

		// Convert each element in the array
		convertedArray := make([]interface{}, len(v))
		for i, item := range v {
			convertedArray[i], _ = convertObjectToValue(item)
		}
		return convertedArray, true
	}

	return convertObjectToValue(result)
}

func handleMetadataValue(value string, exists bool) (interface{}, bool) {
	if !exists {
		return nil, true
	}
	if len(value) == 0 {
		return true, true
	}
	return inferValue(value), true
}

func convertObjectToValue(object interface{}) (interface{}, bool) {
	switch v := object.(type) {
	case bool:
		return v, true
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case string:
		return inferValue(v), true
	}
	return nil, true
}

// inferValue attempts to convert a string to its most appropriate type:
// bool, int, float, date, or keeps it as string if no conversion works
func inferValue(s string) interface{} {
	// Try to parse as boolean
	if strings.ToLower(s) == "true" {
		return true
	}
	if strings.ToLower(s) == "false" {
		return false
	}

	// Try to parse as integer
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return float64(i) // Using float64 for consistency
	}

	// Try to parse as float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	// Try to parse as date (RFC3339 format)
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}

	// Try additional date formats
	dateFormats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006/01/02",
		"01/02/2006",
		time.RFC822,
		time.RFC1123,
	}

	for _, format := range dateFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}

	// Default to string
	return stringValue(s)
}
