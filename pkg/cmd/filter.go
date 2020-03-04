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

	"github.com/yaacov/tree-search-language/v5/pkg/tsl"
	"github.com/yaacov/tree-search-language/v5/pkg/walkers/ident"
	"github.com/yaacov/tree-search-language/v5/pkg/walkers/semantics"
)

// Filter items using namespace and query.
func (o *SQLOptions) Filter(list []unstructured.Unstructured, query string) ([]unstructured.Unstructured, error) {
	var (
		tree      tsl.Node
		err       error
		namespace = o.namespace
	)

	// If we have a query, prepare the search tree.
	if len(query) > 0 {
		tree, err = tsl.ParseTSL(query)
		if err != nil {
			return nil, err
		}

		// Check and replace user identifiers if alias exist.
		tree, err = ident.Walk(tree, o.checkColumnName)
		if err != nil {
			return nil, err
		}
	}

	// Filter items using namespace and query.
	items := []unstructured.Unstructured{}
	for _, item := range list {
		if !o.allNamespaces && item.GetNamespace() != namespace {
			continue
		}

		// If we have a query, check item.
		if len(query) > 0 {
			matchingFilter, err := semantics.Walk(tree, evalFactory(item))
			if err != nil {
				continue
			}
			if !matchingFilter {
				continue
			}
		}

		items = append(items, item)
	}

	return items, nil
}
