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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"

	"github.com/yaacov/kubectl-sql/pkg/client"
	"github.com/yaacov/kubectl-sql/pkg/filter"
	"github.com/yaacov/kubectl-sql/pkg/printers"
)

// Get the resource list.
func (o *SQLOptions) Get(config *rest.Config) error {
	c := client.Config{
		Config: config,
	}

	f := filter.Config{
		CheckColumnName: o.checkColumnName,
		Query:           o.requestedQuery,
		Namespace:       o.namespace,
		AllNamespaces:   o.allNamespaces,
	}

	// Print resources lists.
	for _, r := range o.requestedResources {
		list, err := c.List(r)
		if err != nil {
			return err
		}

		// Filter items by namespace and query.
		filteredList, err := f.Filter(list)
		if err != nil {
			return err
		}

		err = o.Printer(filteredList)
		if err != nil {
			return err
		}
	}

	return nil
}

// checkColumnName checks if a coulumn name has an alias.
func (o *SQLOptions) checkColumnName(s string) (string, error) {
	// Chekc for aliases.
	if v, ok := o.defaultAliases[s]; ok {
		return v, nil
	}

	// If not found in alias table, return the column name unchanged.
	return s, nil
}

// Printer printout a list of items.
func (o *SQLOptions) Printer(items []unstructured.Unstructured) error {
	// Sanity check
	if len(items) == 0 {
		return nil
	}

	p := printers.Config{
		TableFields: o.defaultTableFields,
		Out:         o.Out,
		ErrOut:      o.ErrOut,
	}

	// Print out
	switch o.outputFormat {
	case "yaml":
		return p.YAML(items)
	case "json":
		return p.JSON(items)
	case "name":
		return p.Name(items)
	default:
		p.Table(items)
	}

	return nil
}
