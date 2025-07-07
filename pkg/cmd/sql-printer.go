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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/yaacov/kubectl-sql/pkg/extract"
	"github.com/yaacov/kubectl-sql/pkg/printers"
)

// Printer printout a list of items.
// This function now uses the new data-based approach internally
// while maintaining backward compatibility.
func (o *SQLOptions) Printer(items []unstructured.Unstructured) error {
	return o.PrinterFromData(items)
}

// PrinterFromData prints out evaluated data using the new data-based approach.
// This separates data evaluation from printing, allowing for better performance
// and cleaner separation of concerns.
func (o *SQLOptions) PrinterFromData(items []unstructured.Unstructured) error {
	// Sanity check
	if len(items) == 0 {
		return nil
	}

	p := printers.Config{
		TableFields:   o.defaultTableFields,
		OrderByFields: o.orderByFields,
		Limit:         o.limit,
		Out:           o.Out,
		ErrOut:        o.ErrOut,
		NoHeaders:     o.noHeaders,
	}

	// For yaml, json, and name outputs, we can use items directly
	switch o.outputFormat {
	case "yaml":
		return p.YAML(items)
	case "json":
		return p.JSON(items)
	case "name":
		return p.Name(items)
	default:
		// For table output, we need to evaluate items into table data
		kind := items[0].GetKind()
		fieldNames := o.defaultTableFields.GetFieldNamesForKind(kind)

		// Create converter to transform unstructured objects to table data
		converter := extract.NewUnstructuredToTableConverter(fieldNames)
		rows := converter.ConvertToTableData(items)

		// Print kind information before the table
		if !o.noHeaders {
			fmt.Fprintf(p.Out, "KIND: %s\t", kind)
		}

		return p.Table(rows, fieldNames)
	}
}
