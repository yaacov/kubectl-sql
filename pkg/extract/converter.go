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

package extract

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/yaacov/kubectl-sql/pkg/eval"
)

// UnstructuredToTableConverter converts unstructured Kubernetes objects into table-ready data
type UnstructuredToTableConverter struct {
	// FieldNames contains the list of field names to extract for table columns
	FieldNames []string
}

// NewUnstructuredToTableConverter creates a new converter with the specified field names
func NewUnstructuredToTableConverter(fieldNames []string) *UnstructuredToTableConverter {
	return &UnstructuredToTableConverter{
		FieldNames: fieldNames,
	}
}

// ConvertToTableData converts a list of unstructured items into table-ready data
// Returns: rows ([]map[string]interface{})
func (c *UnstructuredToTableConverter) ConvertToTableData(items []unstructured.Unstructured) []map[string]interface{} {
	if len(items) == 0 {
		return []map[string]interface{}{}
	}

	rows := make([]map[string]interface{}, len(items))

	for i, item := range items {
		evalFunc := eval.EvalFunctionFactory(item)
		fields := make(map[string]interface{})

		for _, fieldName := range c.FieldNames {
			if value, found := evalFunc(fieldName); found {
				fields[fieldName] = value
			} else {
				fields[fieldName] = nil
			}
		}

		rows[i] = fields
	}

	return rows
}
