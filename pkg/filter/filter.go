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

package filter

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/yaacov/tree-search-language/v5/pkg/tsl"
	"github.com/yaacov/tree-search-language/v5/pkg/walkers/ident"
	"github.com/yaacov/tree-search-language/v5/pkg/walkers/semantics"

	"github.com/yaacov/kubectl-sql/pkg/eval"
)

// Config provides information required filter item list by query.
type Config struct {
	Aliases       map[string]string
	Query         string
	Namespace     string
	AllNamespaces bool
}

// checkColumnName checks if a coulumn name has an alias.
func (c *Config) checkColumnName(s string) (string, error) {
	// Chekc for aliases.
	if v, ok := c.Aliases[s]; ok {
		return v, nil
	}

	// If not found in alias table, return the column name unchanged.
	return s, nil
}

// Run filters items using namespace and query.
func (c *Config) Run(list []unstructured.Unstructured) ([]unstructured.Unstructured, error) {
	var (
		tree     tsl.Node
		err      error
		hasQuery = len(c.Query) > 0
	)

	// If we have a query, prepare the search tree.
	if hasQuery {
		tree, err = tsl.ParseTSL(c.Query)
		if err != nil {
			return nil, err
		}

		// Check and replace user identifiers if alias exist.
		tree, err = ident.Walk(tree, c.checkColumnName)
		if err != nil {
			return nil, err
		}
	}

	// Filter items using namespace and query.
	items := []unstructured.Unstructured{}
	for _, item := range list {
		if !c.AllNamespaces && item.GetNamespace() != c.Namespace {
			continue
		}

		// If we have a query, check item.
		if hasQuery {
			matchingFilter, err := semantics.Walk(tree, eval.Factory(item))
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
