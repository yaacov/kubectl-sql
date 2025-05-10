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

package cmd

import (
	"fmt"
	"strings"

	"github.com/yaacov/kubectl-sql/pkg/output"
	"github.com/yaacov/kubectl-sql/pkg/query"
)

// PrintResults outputs the query results in the specified format
func PrintResults(o *SQLOptions, result []map[string]interface{}, queryOptions *query.QueryOptions) error {
	// Check if there are no results
	if len(result) == 0 {
		fmt.Fprintln(o.streams.Out, "No resources found")
		return nil
	}

	// Print results based on output format
	switch o.outputFormat {
	case "json":
		printer := output.NewJSONPrinter().
			WithPrettyPrint(true).
			AddItems(result)
		return printer.Print()

	case "yaml":
		printer := output.NewYAMLPrinter().
			AddItems(result)
		return printer.Print()

	case "table":
		printer := output.NewTablePrinter().
			WithDebugLevel(queryOptions.DebugLevel)

		// Create headers from select fields
		headers := make([]output.Header, 0, len(queryOptions.Select))
		for _, field := range queryOptions.Select {
			// Use alias if provided, otherwise use field name
			displayName := field.Alias
			if field.Alias == "" {
				displayName = strings.Trim(field.Field, ".()")
			}

			// Create a SelectOption with alias (alias fall back to field)
			selectOption := query.SelectOption{
				Field:   field.Field,
				Alias:   displayName,
				Reducer: field.Reducer,
			}
			headers = append(headers, output.Header{
				DisplayName:  displayName,
				JSONPath:     displayName,
				SelectOption: &selectOption,
			})
		}

		printer.WithHeaders(headers...).
			AddItems(result)

		if o.noHeaders {
			printer.WithoutHeaders()
		}

		return printer.Print()

	case "name":
		nameField := "metadata.name"
		if len(queryOptions.Select) > 0 {
			nameField = queryOptions.Select[0].Field
		}

		printer := output.NewNamePrinter().
			WithNameField(nameField).
			WithQueryOptions(queryOptions).
			AddItems(result)

		return printer.Print()

	default:
		return fmt.Errorf("invalid output format: %s", o.outputFormat)
	}
}
