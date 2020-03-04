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
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Printer printout a list of items.
func (o *SQLOptions) Printer(items []unstructured.Unstructured) error {
	// Sanity check
	if len(items) == 0 {
		return nil
	}

	// Print out
	switch o.outputFormat {
	case "yaml":
		return printerYAML(items)
	case "json":
		return printerJSON(items)
	case "name":
		return printerNames(items)
	default:
		o.printerTable(items)
	}

	return nil
}

func printerYAML(items []unstructured.Unstructured) error {
	for _, item := range items {
		yaml, err := yaml.Marshal(item)
		if err != nil {
			return err
		}

		fmt.Printf("\n%+v\n", string(yaml))
	}

	return nil
}

func printerJSON(items []unstructured.Unstructured) error {
	for _, item := range items {
		yaml, err := json.Marshal(item)
		if err != nil {
			return err
		}

		fmt.Printf("\n%+v\n", string(yaml))
	}

	return nil
}

func printerNames(items []unstructured.Unstructured) error {
	for _, item := range items {
		fmt.Printf("%s\n", item.GetName())
	}

	return nil
}

// Get the table column titles and fields for the items.
func (o *SQLOptions) getTableColumns(items []unstructured.Unstructured) []tableField {
	var evalFunc func(string) (interface{}, bool)

	// Get the default template for this kind.
	kind := items[0].GetKind()
	fields, ok := o.defaultTableFields[kind]
	if !ok {
		fields = o.defaultTableFields["other"]
	}

	// Calculte field widths
	for _, item := range items {
		evalFunc = evalFactory(item)

		for i, field := range fields {
			if value, found := evalFunc(field.Name); found && value != nil {
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

func (o *SQLOptions) printerTable(items []unstructured.Unstructured) error {
	var evalFunc func(string) (interface{}, bool)

	// Get table fields for the items.
	fields := o.getTableColumns(items)

	// Pring table head
	fmt.Printf("\nKIND: %s\n", items[0].GetKind())
	for _, field := range fields {
		if field.width > 0 {
			fmt.Printf(field.template, field.Title)
		}
	}
	fmt.Print("\n")

	// Print table rows
	for _, item := range items {
		evalFunc = evalFactory(item)

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

	return nil
}
