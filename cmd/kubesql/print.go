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
	// replace dots with underline
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
			return item.GetLabels()[key[7:]], len(item.GetLabels()[key]) > 0
		}

		if len(key) > 12 && key[:12] == "annotations." {
			return item.GetAnnotations()[key[12:]], len(item.GetLabels()[key]) > 0
		}

		if key == "created" {
			return item.GetCreationTimestamp().Format(time.RFC3339), true
		}

		if key == "deleted" {
			return item.GetDeletionTimestamp().Format(time.RFC3339), true
		}

		// Split the key to work with the `unstructured.Nested` functions.
		keys := strings.Split(key, ".")

		if value, ok, err := unstructured.NestedInt64(item.Object, keys...); ok && err == nil {
			return value, true
		}

		if value, ok, err := unstructured.NestedFloat64(item.Object, keys...); ok && err == nil {
			return value, true
		}

		if value, ok, err := unstructured.NestedString(item.Object, keys...); ok && err == nil {
			return value, true
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

// Get the fields for print for kind
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
				if value, found := evalFunc(field.name); found {
					fmt.Printf(field.template, value)
				} else {
					fmt.Printf(field.template, "")
				}
			}
		}
		fmt.Print("\n")
	}
}
