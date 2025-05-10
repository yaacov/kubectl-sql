package query

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// GetValueByPathString gets a value from an object using a string path
// The path can use dot notation (e.g. "metadata.name") and array indexing (e.g. "spec.containers[0].name")
// If debugLevel > 0, debug information will be printed to stderr
func GetValueByPathString(obj interface{}, path string, debugLevel int) (interface{}, error) {
	if debugLevel > 0 {
		fmt.Fprintf(os.Stderr, "DEBUG: GetValueByPathString input path: %q\n", path)
	}

	// Remove JSONPath notation if present
	path = strings.TrimPrefix(path, "{{")
	path = strings.TrimSuffix(path, "}}")
	path = strings.TrimSpace(path)

	// Remove leading dot if present
	path = strings.TrimPrefix(path, ".")

	if debugLevel > 0 {
		fmt.Fprintf(os.Stderr, "DEBUG: Cleaned path: %q\n", path)
	}

	// Split the path into parts, handling brackets and dots
	// e.g. "spec.containers[0].name" -> ["spec", "containers[0]", "name"]
	var parts []string
	var currentPart strings.Builder
	insideBrackets := false

	for _, char := range path {
		switch char {
		case '[':
			insideBrackets = true
			currentPart.WriteRune(char)
		case ']':
			insideBrackets = false
			currentPart.WriteRune(char)
		case '.':
			if insideBrackets {
				// If inside brackets, the dot is part of the current segment
				currentPart.WriteRune(char)
			} else {
				// If outside brackets, the dot is a separator
				parts = append(parts, currentPart.String())
				currentPart.Reset()
			}
		default:
			currentPart.WriteRune(char)
		}
	}

	// Add the last part if there's anything left
	if currentPart.Len() > 0 {
		parts = append(parts, currentPart.String())
	}

	if debugLevel > 0 {
		debugInfo := map[string]interface{}{
			"originalPath": path,
			"parsedParts":  parts,
		}
		debugJSON, _ := json.MarshalIndent(debugInfo, "", "  ")
		fmt.Fprintf(os.Stderr, "DEBUG: Path parsing results:\n%s\n", string(debugJSON))
	}

	// Forward debug level to getValueByPath if needed
	result, err := getValueByPath(obj, parts, debugLevel)

	if debugLevel > 0 {
		if err != nil {
			fmt.Fprintf(os.Stderr, "DEBUG: GetValueByPathString error: %v\n", err)
		} else {
			resultType := fmt.Sprintf("%T", result)
			resultPreview := "nil"
			if result != nil {
				// Create a preview of the result
				resultJSON, jsonErr := json.Marshal(result)
				if jsonErr == nil {
					if len(resultJSON) > 100 {
						resultPreview = string(resultJSON[:97]) + "..."
					} else {
						resultPreview = string(resultJSON)
					}
				} else {
					resultPreview = fmt.Sprintf("%v", result)
				}
			}

			fmt.Fprintf(os.Stderr, "DEBUG: GetValueByPathString result: type=%s, value=%s\n",
				resultType, resultPreview)
		}
	}

	return result, err
}

// getValueByPath recursively traverses an object following a path
func getValueByPath(obj interface{}, pathParts []string, debugLevel int) (interface{}, error) {
	if debugLevel > 0 {
		fmt.Fprintf(os.Stderr, "DEBUG: getValueByPath processing path parts: %v\n", pathParts)
		if obj != nil {
			fmt.Fprintf(os.Stderr, "DEBUG: Current object type: %T\n", obj)
		} else {
			fmt.Fprintf(os.Stderr, "DEBUG: Current object is nil\n")
		}
	}

	if len(pathParts) == 0 {
		if debugLevel > 0 {
			fmt.Fprintf(os.Stderr, "DEBUG: No more path parts, returning object\n")
		}
		return obj, nil
	}

	if obj == nil {
		if debugLevel > 0 {
			fmt.Fprintf(os.Stderr, "DEBUG: Cannot access %s on nil value\n", strings.Join(pathParts, "."))
		}
		return nil, fmt.Errorf("cannot access %s on nil value", strings.Join(pathParts, "."))
	}

	part := pathParts[0]
	remainingParts := pathParts[1:]

	if debugLevel > 0 {
		fmt.Fprintf(os.Stderr, "DEBUG: Processing part: %q, remaining parts: %v\n", part, remainingParts)
	}

	// Check if part has array indexing notation [i], map key notation [key], or wildcard [*]
	arrayIndex := -1
	mapKey := ""
	isWildcard := false

	// Run regex matchers
	wildcardMatch := regexp.MustCompile(`(.*)\[\*\]$`).FindStringSubmatch(part)
	arrayMatch := regexp.MustCompile(`(.*)\[(\d+)\]$`).FindStringSubmatch(part)
	mapMatch := regexp.MustCompile(`(.*)\[([^\]]+)\]$`).FindStringSubmatch(part)

	// Then use flat if conditions to process matches
	if len(wildcardMatch) == 2 {
		part = wildcardMatch[1]
		isWildcard = true
		if debugLevel > 0 {
			fmt.Fprintf(os.Stderr, "DEBUG: Detected wildcard notation, base part: %q\n", part)
		}
	} else if len(arrayMatch) == 3 {
		part = arrayMatch[1]
		index, err := strconv.Atoi(arrayMatch[2])
		if err != nil {
			if debugLevel > 0 {
				fmt.Fprintf(os.Stderr, "DEBUG: Invalid array index in path: %s\n", part)
			}
			return nil, fmt.Errorf("invalid array index in path: %s", part)
		}
		arrayIndex = index
		if debugLevel > 0 {
			fmt.Fprintf(os.Stderr, "DEBUG: Detected array index notation, base part: %q, index: %d\n", part, arrayIndex)
		}
	} else if len(mapMatch) == 3 {
		part = mapMatch[1]
		mapKey = mapMatch[2]
		// Remove quotes if present
		mapKey = strings.Trim(mapKey, `"'`)
		if debugLevel > 0 {
			fmt.Fprintf(os.Stderr, "DEBUG: Detected map key notation, base part: %q, key: %q\n", part, mapKey)
		}
	}

	switch objTyped := obj.(type) {
	case map[string]interface{}:
		// Get value for current part
		value, exists := objTyped[part]
		if !exists {
			// Don't fail if the part is not found, just return nil
			if debugLevel > 0 {
				fmt.Fprintf(os.Stderr, "DEBUG: Key %q not found in map\n", part)
			}
			return nil, nil
		}

		if debugLevel > 0 {
			fmt.Fprintf(os.Stderr, "DEBUG: Found value for key %q, type: %T\n", part, value)
		}

		// Handle wildcard for arrays
		if isWildcard {
			// Check if the value is an array
			arr, ok := value.([]interface{})
			if !ok {
				if debugLevel > 0 {
					fmt.Fprintf(os.Stderr, "DEBUG: Cannot apply wildcard to non-array value: %s\n", part)
				}
				return nil, fmt.Errorf("cannot apply wildcard to non-array value: %s", part)
			}

			if debugLevel > 0 {
				fmt.Fprintf(os.Stderr, "DEBUG: Processing wildcard on array with %d elements\n", len(arr))
			}

			// For wildcard, collect results from all array elements
			var results []interface{}
			for i, item := range arr {
				if debugLevel > 0 {
					fmt.Fprintf(os.Stderr, "DEBUG: Processing wildcard array element %d\n", i)
				}
				result, err := getValueByPath(item, remainingParts, debugLevel)
				if err == nil && result != nil {
					// Check if result is an array itself and flatten if needed
					if resultArray, isArray := result.([]interface{}); isArray {
						// Flatten by appending individual elements
						if debugLevel > 0 {
							fmt.Fprintf(os.Stderr, "DEBUG: Flattening array result with %d elements\n", len(resultArray))
						}
						results = append(results, resultArray...)
					} else {
						// Non-array result, append as is
						if debugLevel > 0 {
							fmt.Fprintf(os.Stderr, "DEBUG: Adding non-array result to results\n")
						}
						results = append(results, result)
					}
				}
			}
			if debugLevel > 0 {
				fmt.Fprintf(os.Stderr, "DEBUG: Wildcard processing complete, collected %d results\n", len(results))
			}
			return results, nil
		}

		// Handle array indexing if present
		if arrayIndex >= 0 {
			// Check if the value is an array
			if arr, ok := value.([]interface{}); ok {
				if arrayIndex >= len(arr) {
					if debugLevel > 0 {
						fmt.Fprintf(os.Stderr, "DEBUG: Array index out of bounds: %d (array length: %d)\n", arrayIndex, len(arr))
					}
					return nil, fmt.Errorf("array index out of bounds: %d", arrayIndex)
				}
				if debugLevel > 0 {
					fmt.Fprintf(os.Stderr, "DEBUG: Accessing array element at index %d\n", arrayIndex)
				}
				value = arr[arrayIndex]
			} else {
				if debugLevel > 0 {
					fmt.Fprintf(os.Stderr, "DEBUG: Cannot apply array index to non-array value: %s\n", part)
				}
				return nil, fmt.Errorf("cannot apply array index to non-array value: %s", part)
			}
		}

		// Handle map key access if present
		if mapKey != "" {
			// Check if the value is a map
			if m, ok := value.(map[string]interface{}); ok {
				mapValue, exists := m[mapKey]
				if !exists {
					if debugLevel > 0 {
						fmt.Fprintf(os.Stderr, "DEBUG: Map key %q not found\n", mapKey)
					}
					return nil, nil
				}
				if debugLevel > 0 {
					fmt.Fprintf(os.Stderr, "DEBUG: Accessing map with key %q\n", mapKey)
				}
				value = mapValue
			} else {
				if debugLevel > 0 {
					fmt.Fprintf(os.Stderr, "DEBUG: Cannot apply map key to non-map value: %s\n", part)
				}
				return nil, fmt.Errorf("cannot apply map key to non-map value: %s", part)
			}
		}

		// If this is the last part, return the value
		if len(remainingParts) == 0 {
			if debugLevel > 0 {
				fmt.Fprintf(os.Stderr, "DEBUG: Reached end of path, returning final value of type %T\n", value)
			}
			return value, nil
		}

		// Otherwise, continue recursing
		if debugLevel > 0 {
			fmt.Fprintf(os.Stderr, "DEBUG: Continuing recursion with remaining parts\n")
		}
		return getValueByPath(value, remainingParts, debugLevel)

	default:
		if debugLevel > 0 {
			fmt.Fprintf(os.Stderr, "DEBUG: Cannot access property %s on non-object value of type %T\n", part, obj)
		}
		return nil, fmt.Errorf("cannot access property %s on non-object value", part)
	}
}

// GetValue retrieves a value by name (or alias) from obj using JSONPath,
// then applies a reducer if one is specified in queryOpts.
func GetValue(obj interface{}, name string, queryOpts *QueryOptions) (interface{}, error) {
	var reducer string
	var matchedSelectOpt *SelectOption

	// We store sanitied version of the name in the alias field
	sanitizedName := strings.Trim(name, ".()")
	path := sanitizedName

	// If name matches an alias, switch to its Field and capture reducer
	for _, opt := range queryOpts.Select {
		if opt.Alias == sanitizedName {
			path = opt.Field
			reducer = opt.Reducer
			matchedSelectOpt = &opt
			break
		}
	}

	// Check for debug level from QueryOptions
	debugLevel := queryOpts.DebugLevel

	// Check if path is a default path alias
	path = GetDefaultFieldAlias(path)

	// Print consolidated debug info if debug level is enabled
	if debugLevel > 0 {
		debugInfo := map[string]interface{}{
			"input": map[string]interface{}{
				"name":        name,
				"sanitized":   sanitizedName,
				"resultPath":  path,
				"hasReducer":  reducer != "",
				"reducerType": reducer,
			},
			"matchFound":        matchedSelectOpt != nil,
			"finalResolvedPath": path,
		}

		// Add matched select option information if available
		if matchedSelectOpt != nil {
			debugInfo["matchedSelectOption"] = *matchedSelectOpt
		} else {
			debugInfo["note"] = "No select option matched, using original path"
		}

		debugJSON, _ := json.MarshalIndent(debugInfo, "", "  ")
		fmt.Fprintf(os.Stderr, "DEBUG: GetValue operation:\n%s\n", string(debugJSON))
	}

	// Sanity check for empty path
	if path == "" {
		return nil, fmt.Errorf("no path found for name %q", name)
	}

	// Fetch the raw value
	val, err := GetValueByPathString(obj, path, debugLevel)
	if err != nil {
		return nil, err
	}

	// Debug if val is empty
	if debugLevel > 0 {
		isEmpty := false
		var valueType string

		if val == nil {
			isEmpty = true
			valueType = "nil"
		} else {
			switch v := val.(type) {
			case string:
				isEmpty = v == ""
				valueType = "string"
			case []interface{}:
				isEmpty = len(v) == 0
				valueType = "array"
			case map[string]interface{}:
				isEmpty = len(v) == 0
				valueType = "map"
			default:
				valueType = fmt.Sprintf("%T", val)
			}
		}

		if isEmpty {
			fmt.Fprintf(os.Stderr, "DEBUG: Value for path %q is empty (type: %s)\n", path, valueType)
		} else {
			fmt.Fprintf(os.Stderr, "DEBUG: Value for path %q is non-empty (type: %s)\n", path, valueType)
		}
	}

	// Check is it's a string array, and apply ParseString if so
	if parsedArr, ok := val.([]interface{}); ok {
		for i, v := range parsedArr {
			if str, ok := v.(string); ok {
				parsed, err := ParseString(str)
				if err != nil {
					return nil, fmt.Errorf("failed to parse string %q: %v", str, err)
				}
				parsedArr[i] = parsed
			}
		}

		val = parsedArr
	}

	// If single string, try to parse as a more specific type
	if str, ok := val.(string); ok {
		parsed, _ := ParseString(str)

		val = parsed
	}

	// Apply reducer if specified
	if reducer != "" {
		if debugLevel > 0 {
			fmt.Fprintf(os.Stderr, "DEBUG: Applying reducer %q to value %+v (type: %T)\n", reducer, val, val)
		}

		// Check if val is a slice
		if _, ok := val.([]interface{}); !ok {
			return nil, fmt.Errorf("reducer %q can only be applied to array values", reducer)
		}

		// Apply the reducer
		reduced, err := applyReducer(val, reducer)
		if err != nil {
			return nil, fmt.Errorf("failed to apply reducer %q: %v", reducer, err)
		}
		return reduced, nil
	}

	return val, nil
}
