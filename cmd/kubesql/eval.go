// Copyright 2020 Yaacov Zamir <kobi.zamir@gmail.com>
// and other contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Author: 2020 Yaacov Zamir <kobi.zamir@gmail.com>

// Package main.
package main

import (
	"math"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/yaacov/tree-search-language/v5/pkg/walkers/semantics"

	"github.com/urfave/cli/v2"
)

// checkColumnName checks if a coulumn name has an alias.
func checkColumnName(s string) (string, error) {
	// Chekc for aliases.
	if v, ok := defaultAliases[s]; ok {
		return v, nil
	}

	// If not found in alias table, return the column name unchanged.
	return s, nil
}

// Retrun a nested object using a key
func getNestedObject(object interface{}, key string) (interface{}, bool) {
	keys := strings.Split(key, ".")

	var objectList []interface{}
	var objectMap map[string]interface{}
	ok := true

	for _, key := range keys {
		if i, err := strconv.ParseUint(key, 10, 64); err == nil && i > 0 {
			if objectList, ok = object.([]interface{}); !ok {
				break
			}

			if ok = i <= uint64(len(objectList)); !ok {
				break
			}

			object = objectList[i-1]
		} else {
			if objectMap, ok = object.(map[string]interface{}); !ok {
				break
			}

			if object, ok = objectMap[key]; !ok {
				break
			}
		}
	}

	return object, ok
}

func stringValue(str string, verbose bool) interface{} {
	// Check for SI numbers
	multiplier := 0.0
	s := str

	// Remove SI `i` if exist
	// Note: we support "K", "M", "G" and "Ki", "Mi", "Gi" postfix
	if len(s) > 1 && s[len(s)-1:] == "i" {
		s = s[:len(s)-1]
	}

	// Check for SI postfix
	if len(s) > 1 {
		postfix := s[len(s)-1:]
		switch postfix {
		case "K":
			multiplier = 1024.0
		case "M":
			multiplier = math.Pow(1024, 2)
		case "G":
			multiplier = math.Pow(1024, 3)
		case "T":
			multiplier = math.Pow(1024, 4)
		case "P":
			multiplier = math.Pow(1024, 5)
		}

		if multiplier >= 1.0 {
			s = s[:len(s)-1]
		}

		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			newValue := float64(i) * multiplier
			debugLog(verbose, "converting units, %v (%f)\n", s, newValue)

			return newValue
		}
	}

	// Check for false / true
	if str == "true" || str == "True" {
		debugLog(verbose, "converting type to boolean %v\n", str)
		return true
	}
	if str == "false" || str == "False" {
		debugLog(verbose, "converting type to boolean %v\n", str)
		return false
	}

	// Check for RFC3339 dates
	if t, err := time.Parse(time.RFC3339, str); err == nil {
		return t
	}

	// Default to string
	debugLog(verbose, "default to string %v\n", str)
	return str
}

// evalFactory extract a value from an item using a key.
func evalFactory(c *cli.Context, item unstructured.Unstructured) semantics.EvalFunc {
	return func(key string) (interface{}, bool) {
		verbose := c.Bool("verbose")

		// Check for reserved words.
		switch key {
		case "name":
			return item.GetName(), true
		case "namespace":
			return item.GetNamespace(), true
		case "created":
			return item.GetCreationTimestamp().Time, true
		case "deleted":
			return item.GetDeletionTimestamp().Time, true
		}

		// Check for labels and annotations.
		if len(key) > 7 && key[:7] == "labels." {
			value, ok := item.GetLabels()[key[7:]]

			debugLog(verbose, "parse labels for key, %v\n", key[7:])

			// Empty label represent the label is present
			if ok && len(value) == 0 {
				return true, true
			}

			// Missing value
			if !ok {
				return nil, true
			}

			v := stringValue(value, verbose)
			return v, true
		}

		if len(key) > 12 && key[:12] == "annotations." {
			value, ok := item.GetAnnotations()[key[12:]]

			debugLog(verbose, "parse annotations for key, %v\n", key[12:])

			// Empty annotations represent the annotations is present
			if ok && len(value) == 0 {
				return true, true
			}

			// Missing value
			if !ok {
				return nil, true
			}

			v := stringValue(value, verbose)
			return v, true
		}

		// Check for numbers, booleans, dates and strings.
		object, ok := getNestedObject(item.Object, key)
		if !ok {
			debugLog(verbose, "failed to find an object for key, %v\n", key)

			// Missing value is interpated as null value.
			return nil, true
		}

		switch object.(type) {
		case float64:
			return object.(float64), true
		case int64:
			return float64(object.(int64)), true
		case string:
			v := stringValue(object.(string), verbose)

			return v, true
		}

		// Missing value is interpated as null value.
		debugLog(verbose, "failed to parse value, %v\n", object)
		return nil, true
	}
}
