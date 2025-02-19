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

package printers

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/yaacov/kubectl-sql/pkg/eval"
)

// TableField describes how to print the SQL results table.
type TableField struct {
	Title    string `json:"title"`
	Name     string `json:"name"`
	Width    int
	Template string
}
type tableFields []TableField

// TableFieldsMap a map of lists of table field descriptions.
type TableFieldsMap map[string]tableFields

// Config provides information required filter item list by query.
type Config struct {
	// TableFields describe table field columns
	TableFields TableFieldsMap
	// Out think, os.Stdout
	Out io.Writer
	// ErrOut think, os.Stderr
	ErrOut io.Writer
}

// Get the table column titles and fields for the items.
func (c *Config) getTableColumns(items []unstructured.Unstructured) tableFields {
	var evalFunc func(string) (interface{}, bool)

	// Get the default template for this kind.
	kind := items[0].GetKind()

	// Try different variations of kind name
	fields, ok := c.TableFields[kind]
	if !ok {
		fields, ok = c.TableFields[strings.ToLower(kind)]
		if !ok {
			fields, ok = c.TableFields[strings.ToLower(kind)+"s"]
			if !ok {
				fields = c.TableFields["other"]
			}
		}
	}

	// Zero out field width
	for i := range fields {
		fields[i].Width = 0
		fields[i].Template = ""
	}

	// Calculte field widths
	for _, item := range items {
		evalFunc = eval.EvalFunctionFactory(item)

		for i, field := range fields {
			if value, found := evalFunc(field.Name); found && value != nil {
				length := len(fmt.Sprintf("%v", value))

				if length > fields[i].Width {
					fields[i].Width = length
				}
			}
		}
	}

	// Calculte field template
	for i, field := range fields {
		if field.Width > 0 {
			// Ajdust for title length
			width := len(field.Title)
			if width < field.Width {
				width = field.Width
			}

			fields[i].Template = fmt.Sprintf("%%-%ds\t", width)
		}
	}

	return fields
}

// Table prints items in Table format
func (c *Config) Table(items []unstructured.Unstructured) error {
	var evalFunc func(string) (interface{}, bool)

	// Get table fields for the items.
	fields := c.getTableColumns(items)

	// Print table head
	fmt.Fprintf(c.Out, "KIND: %s\tCOUNT: %d\n", items[0].GetKind(), len(items))
	for _, field := range fields {
		if field.Width > 0 {
			fmt.Fprintf(c.Out, field.Template, field.Title)
		}
	}
	fmt.Print("\n")

	// Print table rows
	for _, item := range items {
		evalFunc = eval.EvalFunctionFactory(item)

		for _, field := range fields {
			if field.Width > 0 {
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

					fmt.Fprintf(c.Out, field.Template, value)
				} else {
					fmt.Fprintf(c.Out, field.Template, "")
				}
			}
		}
		fmt.Print("\n")
	}

	return nil
}
