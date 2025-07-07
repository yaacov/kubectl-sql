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
	"time"
)

// TableField describes how to print the SQL results table.
type TableField struct {
	Title    string `json:"title"`
	Name     string `json:"name"`
	Width    int
	Template string
}

// TableFields is a slice of TableField
type TableFields []TableField

// TableFieldsMap a map of lists of table field descriptions.
type TableFieldsMap map[string]TableFields

// Config provides information required filter item list by query.
type Config struct {
	// TableFields describe table field columns
	TableFields TableFieldsMap
	// OrderByFields describes how to sort the table results
	OrderByFields []OrderByField
	// Limit restricts the number of results displayed (0 means no limit)
	Limit int
	// NoHeaders if true, don't print header rows
	NoHeaders bool
	// Out think, os.Stdout
	Out io.Writer
	// ErrOut think, os.Stderr
	ErrOut io.Writer
}

const (
	// SelectedFields is used to identify fields specifically selected in a SQL query
	SelectedFields = "selected"
)

// ExtractFieldNames extracts field names from TableField configurations
func (tf TableFields) ExtractFieldNames() []string {
	fieldNames := make([]string, len(tf))
	for i, field := range tf {
		fieldNames[i] = field.Name
	}
	return fieldNames
}

// GetFieldNamesForKind returns the appropriate field names for a given kind
func (tfm TableFieldsMap) GetFieldNamesForKind(kind string) []string {
	// Try different variations of kind name
	fields, ok := tfm[SelectedFields]
	if !ok || fields == nil {
		fields, ok = tfm[kind]
		if !ok || fields == nil {
			fields = tfm["other"]
		}
	}

	return fields.ExtractFieldNames()
}

// GetTableFieldsForKind returns the appropriate table fields for a given kind
func (tfm TableFieldsMap) GetTableFieldsForKind(kind string) TableFields {
	// Try different variations of kind name
	fields, ok := tfm[SelectedFields]
	if !ok || fields == nil {
		fields, ok = tfm[kind]
		if !ok || fields == nil {
			fields = tfm["other"]
		}
	}

	return fields
}

// getTableColumnsFromData calculates table column widths and templates from evaluated data
func (c *Config) getTableColumnsFromData(rows []map[string]interface{}, fieldNames []string) TableFields {
	// Create fields based on the field names in the table data
	fields := make(TableFields, len(fieldNames))
	for i, fieldName := range fieldNames {
		fields[i] = TableField{
			Title: fieldName,
			Name:  fieldName,
			Width: 0,
		}
	}

	// Zero out field width
	for i := range fields {
		fields[i].Width = 0
		fields[i].Template = ""
	}

	// Calculate field widths from the evaluated data
	for _, row := range rows {
		for i, field := range fields {
			if value, found := row[field.Name]; found && value != nil {
				length := len(fmt.Sprintf("%v", value))

				if length > fields[i].Width {
					fields[i].Width = length
				}
			}
		}
	}

	// Calculate field template
	for i, field := range fields {
		if field.Width > 0 {
			// Adjust for title length
			width := len(field.Title)
			if width < field.Width {
				width = field.Width
			}

			fields[i].Template = fmt.Sprintf("%%-%ds\t", width)
		}
	}

	return fields
}

// Table prints evaluated data in table format
func (c *Config) Table(rows []map[string]interface{}, fieldNames []string) error {
	if len(rows) == 0 {
		return nil
	}

	// Get table fields for the items using the evaluated data
	fields := c.getTableColumnsFromData(rows, fieldNames)

	// Apply limit if set
	displayCount := len(rows)
	if c.Limit > 0 && c.Limit < displayCount {
		displayCount = c.Limit
	}

	// Print table head if headers are not disabled
	if !c.NoHeaders {
		total := len(rows)
		fmt.Fprintf(c.Out, "COUNT: %d", total)
		if c.Limit > 0 && c.Limit < total {
			fmt.Fprintf(c.Out, "\tDISPLAYING: %d", displayCount)
		}
		fmt.Fprintf(c.Out, "\n")

		for _, field := range fields {
			if field.Width > 0 {
				fmt.Fprintf(c.Out, field.Template, field.Title)
			}
		}
		fmt.Print("\n")
	}

	// Print table rows
	for i, row := range rows {
		// Respect the limit if set
		if c.Limit > 0 && i >= c.Limit {
			break
		}

		for _, field := range fields {
			if field.Width > 0 {
				if v, found := row[field.Name]; found && v != nil {
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
