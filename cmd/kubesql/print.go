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
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/yaacov/tree-search-language/v5/pkg/tsl"
	"github.com/yaacov/tree-search-language/v5/pkg/walkers/ident"
	"github.com/yaacov/tree-search-language/v5/pkg/walkers/semantics"

	"github.com/urfave/cli/v2"
)

// tableField describes how to print a table column.
type tableField struct {
	title    string
	name     string
	width    int
	template string
}

// checkColumnName checks if a coulumn name is valid in user space replace it
// with the mapped column name and returns and error if not a valid name.
func checkColumnName(s string) (string, error) {
	// replace aliases with complete resource path.
	return s, nil
}

// evalFactory creates an eval function that extract a value from an item using a key.
func evalFactory(c *cli.Context, item unstructured.Unstructured) semantics.EvalFunc {
	return func(key string) (interface{}, bool) {
		if key == "name" {
			return item.GetName(), true
		}

		if key == "namespace" {
			return item.GetNamespace(), true
		}

		if len(key) > 7 && key[:7] == "labels." {
			value, ok := item.GetLabels()[key[7:]]

			// Empty label represent the label is present
			if ok && len(value) == 0 {
				value = "true"
			}

			// Missing value
			if !ok {
				return nil, true
			}

			return value, ok
		}

		if len(key) > 12 && key[:12] == "annotations." {
			value, ok := item.GetLabels()[key[12:]]

			// Empty annotations represent the annotations is present
			if ok && len(value) == 0 {
				value = "true"
			}

			// Missing value
			if !ok {
				return nil, true
			}

			return value, ok
		}

		if key == "created" {
			return item.GetCreationTimestamp().Format(time.RFC3339), true
		}

		if key == "deleted" {
			return item.GetDeletionTimestamp().Format(time.RFC3339), true
		}

		// Split the key to work with the `unstructured.Nested` functions.
		keys := strings.Split(key, ".")

		var object interface{}
		var objectList []interface{}
		var objectMap map[string]interface{}
		ok := true
		object = item.Object

		for _, key := range keys {
			if i, err := strconv.ParseUint(key, 10, 64); err == nil && i > 0 {
				objectList, ok = object.([]interface{})
				if !ok {
					break
				}

				ok = i < uint64(len(objectList))
				if !ok {
					break
				}

				object = objectList[i]
			} else {
				objectMap, ok = object.(map[string]interface{})
				if !ok {
					break
				}

				object, ok = objectMap[key]
				if !ok {
					break
				}
			}
		}

		if ok {
			switch object.(type) {
			case float64:
				return object.(float64), true
			case int64:
				return float64(object.(int64)), true
			case string:
				if c.Bool("si-units") {
					multiplier := 0.0
					s := object.(string)

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

							if i, err := strconv.ParseInt(s, 10, 64); err == nil {
								newValue := float64(i) * multiplier
								if c.Bool("verbose") {
									log.Printf("converting units, %v (%f)\n", object, newValue)
								}

								return newValue, true
							}
						}
					}
				}
				return object.(string), true
			}
		}

		if c.Bool("verbose") {
			v, found, err := unstructured.NestedFieldNoCopy(item.Object, keys...)

			log.Printf("failed to parse query value, %v (%v) - %v\n", v, found, err)
		}

		// Missing value is interpated as null value.
		return nil, true
	}
}

// Use namespace and query to printout items.
func printItems(c *cli.Context, list *unstructured.UnstructuredList, namespace string, query string) {
	var (
		tree tsl.Node
		err  error
	)

	// If we have a query, prepare the search tree.
	if len(query) > 0 {
		tree, err = tsl.ParseTSL(query)
		errExit("Failed to parse query", err)

		// Check and replace user identifiers with the document field names.
		tree, err = ident.Walk(tree, checkColumnName)
		errExit("Failed to parse query itentifiers", err)
	}

	items := []unstructured.Unstructured{}
	for _, item := range list.Items {
		if namespace != "" && item.GetNamespace() != namespace {
			continue
		}

		// If we have a query, check item.
		if len(query) > 0 {
			matchingFilter, err := semantics.Walk(tree, evalFactory(c, item))
			if err != nil {
				if c.Bool("verbose") {
					log.Printf("failed to query item: %v", err)
				}
				continue
			}
			if !matchingFilter {
				continue
			}
		}

		items = append(items, item)
	}

	// Check for items
	if len(items) == 0 {
		if c.Bool("verbose") {
			log.Print("no matching items found")
		}
		return
	}

	// Print out
	switch c.String("output") {
	case "yaml":
		printItemsYAML(items)
	case "json":
		printItemsJSON(items)
	default:
		printItemsTable(c, items)
	}
}

func printItemsYAML(items []unstructured.Unstructured) {
	for _, item := range items {
		yaml, err := yaml.Marshal(item)
		errExit("Failed to marshal item", err)

		fmt.Printf("\n%+v\n", string(yaml))
	}
}

func printItemsJSON(items []unstructured.Unstructured) {
	for _, item := range items {
		yaml, err := json.Marshal(item)
		errExit("Failed to marshal item", err)

		fmt.Printf("\n%+v\n", string(yaml))
	}
}

// Get the table column titles and fields by resource kind.
func getFields(kind string) []tableField {
	switch kind {
	case "Pod":
		return []tableField{
			{
				title: "NAMESPACE",
				name:  "namespace",
			},
			{
				title: "NAME",
				name:  "name",
			},
			{
				title: "PHASE",
				name:  "status.phase",
			},
			{
				title: "hostIP",
				name:  "status.hostIP",
			},
			{
				title: "CREATION_TIME(RFC3339)",
				name:  "created",
			},
		}
	default:
		return []tableField{
			{
				title: "NAMESPACE",
				name:  "namespace",
			},
			{
				title: "NAME",
				name:  "name",
			},
			{
				title: "CREATION_TIME(RFC3339)",
				name:  "created",
			},
		}
	}
}

func printItemsTable(c *cli.Context, items []unstructured.Unstructured) {
	var evalFunc func(string) (interface{}, bool)

	// Get table fields for this item.
	kind := items[0].GetKind()
	fields := getFields(kind)

	// Calculte field widths
	for _, item := range items {
		evalFunc = evalFactory(c, item)

		for i, field := range fields {
			if value, found := evalFunc(field.name); found {
				length := len(fmt.Sprintf("%v", value))

				if length > fields[i].width {
					fields[i].width = length
				}
			}
		}
	}

	// Calculte field template
	for i, field := range fields {
		if field.width > 0 {
			// Ajdust for title length
			width := len(field.title)
			if width < field.width {
				width = field.width
			}

			fields[i].template = fmt.Sprintf("%%-%ds\t", width)
		}
	}

	// Pring table head
	for _, field := range fields {
		if field.width > 0 {
			fmt.Printf(field.template, field.title)
		}
	}
	fmt.Print("\n")

	// Print table rows
	for _, item := range items {
		evalFunc = evalFactory(c, item)

		for _, field := range fields {
			if field.width > 0 {
				if value, found := evalFunc(field.name); found && value != nil {
					fmt.Printf(field.template, value)
				} else {
					fmt.Printf(field.template, "")
				}
			}
		}
		fmt.Print("\n")
	}
}
