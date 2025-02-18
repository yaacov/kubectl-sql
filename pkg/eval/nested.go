package eval

import (
	"strconv"
	"strings"
)

// getNestedObject returns a nested object using a dot-separated key path
func getNestedObject(object interface{}, key string) (interface{}, bool) {
	if key == "" {
		return object, true
	}

	keys := strings.Split(key, ".")
	return getNestedValue(object, keys)
}

// getNestedValue recursively traverses the object using the keys array
func getNestedValue(obj interface{}, keys []string) (interface{}, bool) {
	if len(keys) == 0 {
		return obj, true
	}

	currentKey := keys[0]
	remainingKeys := keys[1:]

	// Handle array access with index notation (e.g., "items[1]")
	if name, index, isArray := parseArrayIndex(currentKey); isArray {
		// Get the array first
		m, ok := obj.(map[string]interface{})
		if !ok {
			return nil, false
		}

		array, exists := m[name]
		if !exists {
			return nil, false
		}

		// Get the array element
		list, ok := array.([]interface{})
		if !ok || index > uint64(len(list)) {
			return nil, false
		}

		return getNestedValue(list[index-1], remainingKeys)
	}

	// Handle numeric indices (e.g., "items.1")
	if index, err := strconv.ParseUint(currentKey, 10, 64); err == nil && index > 0 {
		list, ok := obj.([]interface{})
		if !ok || index > uint64(len(list)) {
			return nil, false
		}

		return getNestedValue(list[index-1], remainingKeys)
	}

	// Handle regular map access
	m, ok := obj.(map[string]interface{})
	if !ok {
		return nil, false
	}

	val, exists := m[currentKey]
	if !exists {
		return nil, false
	}

	return getNestedValue(val, remainingKeys)
}

// parseArrayIndex extracts array name and index from a string like "name[index]"
func parseArrayIndex(key string) (name string, index uint64, isArrayAccess bool) {
	arrayIndex := strings.LastIndex(key, "[")
	if arrayIndex == -1 || !strings.HasSuffix(key, "]") {
		return "", 0, false
	}

	indexStr := key[arrayIndex+1 : len(key)-1]
	name = key[:arrayIndex]

	// Convert index string to uint64
	i, err := strconv.ParseUint(indexStr, 10, 64)
	if err != nil || i == 0 {
		return "", 0, false
	}

	return name, i, true
}
