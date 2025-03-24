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
	"reflect"
	"sort"
	"strconv"
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
	// OrderByFields describes how to sort the table results
	OrderByFields []OrderByField
	// Limit restricts the number of results displayed (0 means no limit)
	Limit int
	// Out think, os.Stdout
	Out io.Writer
	// ErrOut think, os.Stderr
	ErrOut io.Writer
}

const (
	// SelectedFields is used to identify fields specifically selected in a SQL query
	SelectedFields = "selected"
)

// Get the table column titles and fields for the items.
func (c *Config) getTableColumns(items []unstructured.Unstructured) tableFields {
	var evalFunc func(string) (interface{}, bool)

	// Get the default template for this kind.
	kind := items[0].GetKind()

	// Try different variations of kind name
	fields, ok := c.TableFields[SelectedFields]
	if !ok || fields == nil {
		fields, ok = c.TableFields[kind]
		if !ok || fields == nil {
			fields = c.TableFields["other"]
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

// sortItems sorts the slice of unstructured items based on the OrderByFields
func (c *Config) sortItems(items []unstructured.Unstructured) {
	if len(c.OrderByFields) == 0 {
		return
	}

	sort.SliceStable(items, func(i, j int) bool {
		for _, orderBy := range c.OrderByFields {
			evalFuncI := eval.EvalFunctionFactory(items[i])
			evalFuncJ := eval.EvalFunctionFactory(items[j])

			valueI, foundI := evalFuncI(orderBy.Name)
			valueJ, foundJ := evalFuncJ(orderBy.Name)

			// If either value is not found, prioritize the found value
			if !foundI && foundJ {
				return !orderBy.Descending
			}
			if foundI && !foundJ {
				return orderBy.Descending
			}
			if !foundI && !foundJ {
				continue
			}

			// Both values found, compare them
			if valueI == nil && valueJ != nil {
				return !orderBy.Descending
			}
			if valueI != nil && valueJ == nil {
				return orderBy.Descending
			}

			// Compare values
			switch vI := valueI.(type) {
			case bool:
				vJ := valueJ.(bool)
				if vI != vJ {
					return vI != orderBy.Descending
				}
			case float64:
				vJ := valueJ.(float64)
				if vI != vJ {
					return vI < vJ != orderBy.Descending
				}
			case string:
				vJ := valueJ.(string)
				if vI != vJ {
					return vI < vJ != orderBy.Descending
				}
			case time.Time:
				vJ := valueJ.(time.Time)
				if !vI.Equal(vJ) {
					return vI.Before(vJ) != orderBy.Descending
				}
			default:
				// Fallback to reflect.DeepEqual for other types
				if !reflect.DeepEqual(valueI, valueJ) {
					return reflect.DeepEqual(valueI, valueJ) != orderBy.Descending
				}
			}
		}
		return false
	})
}

// Table prints items in Table format
func (c *Config) Table(items []unstructured.Unstructured) error {
	var evalFunc func(string) (interface{}, bool)

	// Sort items if OrderByFields is set
	c.sortItems(items)

	// Get table fields for the items.
	fields := c.getTableColumns(items)

	// Apply limit if set
	displayCount := len(items)
	if c.Limit > 0 && c.Limit < displayCount {
		displayCount = c.Limit
	}

	// Print table head
	fmt.Fprintf(c.Out, "KIND: %s\tCOUNT: %d", items[0].GetKind(), len(items))
	if c.Limit > 0 && c.Limit < len(items) {
		fmt.Fprintf(c.Out, "\tDISPLAYING: %d", displayCount)
	}
	fmt.Fprintf(c.Out, "\n")

	for _, field := range fields {
		if field.Width > 0 {
			fmt.Fprintf(c.Out, field.Template, field.Title)
		}
	}
	fmt.Print("\n")

	// Print table rows
	for i, item := range items {
		// Respect the limit if set
		if c.Limit > 0 && i >= c.Limit {
			break
		}

		evalFunc = eval.EvalFunctionFactory(item)

		for _, field := range fields {
			if field.Width > 0 {
				if v, found := evalFunc(field.Name); found && v != nil {
					value := v
					switch v := v.(type) {
					case bool:
						value = "false"
						if v {
							value = "true"
						}
					case float64:
						value = strconv.FormatFloat(v, 'f', -1, 64)
					case time.Time:
						value = v.Format(time.RFC3339)
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
