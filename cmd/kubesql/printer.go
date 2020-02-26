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
	"strconv"
	"time"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/urfave/cli/v2"
)

// Use namespace and query to printout items.
func printer(c *cli.Context, list *unstructured.UnstructuredList, namespace string, query string) {
	verbose := c.Bool("verbose")

	// Filter items using namespace and query.
	items := filter(c, list, namespace, query)

	// Sanity check
	if len(items) == 0 {
		debugLog(verbose, "no matching items found")
		return
	}

	// Print out
	switch c.String("output") {
	case "yaml":
		printerYAML(items)
	case "json":
		printerJSON(items)
	case "name":
		printerNames(items)
	default:
		printerTable(c, items)
	}
}

func printerYAML(items []unstructured.Unstructured) {
	for _, item := range items {
		yaml, err := yaml.Marshal(item)
		errExit("Failed to marshal item", err)

		fmt.Printf("\n%+v\n", string(yaml))
	}
}

func printerJSON(items []unstructured.Unstructured) {
	for _, item := range items {
		yaml, err := json.Marshal(item)
		errExit("Failed to marshal item", err)

		fmt.Printf("\n%+v\n", string(yaml))
	}
}

func printerNames(items []unstructured.Unstructured) {
	for _, item := range items {
		fmt.Printf("%s\n", item.GetName())
	}
}

// Get the table column titles and fields for the items.
func getTableColumns(c *cli.Context, items []unstructured.Unstructured) []tableField {
	var evalFunc func(string) (interface{}, bool)

	// Get the default template for this kind.
	kind := items[0].GetKind()
	fields, ok := defaultTableFields[kind]
	if !ok {
		fields = defaultTableFields["other"]
	}

	// Calculte field widths
	for _, item := range items {
		evalFunc = evalFactory(c, item)

		for i, field := range fields {
			if value, found := evalFunc(field.Name); found {
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
			width := len(field.Title)
			if width < field.width {
				width = field.width
			}

			fields[i].template = fmt.Sprintf("%%-%ds\t", width)
		}
	}

	return fields
}

func printerTable(c *cli.Context, items []unstructured.Unstructured) {
	var evalFunc func(string) (interface{}, bool)
	verbose := c.Bool("verbose")

	// Get table fields for the items.
	fields := getTableColumns(c, items)

	debugLog(verbose, "printing table, %v items %v fields\n", len(items), len(fields))

	// Pring table head
	for _, field := range fields {
		if field.width > 0 {
			fmt.Printf(field.template, field.Title)
		}
	}
	fmt.Print("\n")

	// Print table rows
	for _, item := range items {
		evalFunc = evalFactory(c, item)

		for _, field := range fields {
			if field.width > 0 {
				if v, found := evalFunc(field.Name); found && v != nil {
					value := v
					switch v.(type) {
					case bool:
						value = "false"
						if v.(bool) {
							value = "true"
						}
					case float64:
						value = strconv.FormatFloat(v.(float64), 'f', -1, 64)
					case time.Time:
						value = v.(time.Time).Format(time.RFC3339)
					}

					fmt.Printf(field.template, value)
				} else {
					fmt.Printf(field.template, "")
				}
			}
		}
		fmt.Print("\n")
	}
}
