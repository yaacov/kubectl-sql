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
	"log"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/yaacov/tree-search-language/v5/pkg/tsl"
	"github.com/yaacov/tree-search-language/v5/pkg/walkers/ident"
	"github.com/yaacov/tree-search-language/v5/pkg/walkers/semantics"

	"github.com/urfave/cli/v2"
)

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

		v, found, err := unstructured.NestedFieldNoCopy(item.Object, keys...)

		if c.Bool("verbose") {
			log.Printf("Failed to parse query value, %v (%v) - %v\n", v, found, err)
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

	for _, item := range list.Items {
		if namespace != "" && item.GetNamespace() != namespace {
			continue
		}

		// If we have a query, check item.
		if len(query) > 0 {
			matchingFilter, err := semantics.Walk(tree, evalFactory(c, item))
			if err != nil {
				if c.Bool("verbose") {
					log.Printf("Failed to query item: %v", err)
				}
				continue
			}
			if !matchingFilter {
				continue
			}
		}

		switch c.String("output") {
		case "yaml":
			printItemYAML(item)
		case "json":
			printItemJSON(item)
		default:
			printItemTableRaw(item)
		}
	}
}
