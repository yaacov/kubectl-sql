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

	"github.com/yaacov/tree-search-language/v6/pkg/tsl"
	"github.com/yaacov/tree-search-language/v6/pkg/walkers/ident"
	"github.com/yaacov/tree-search-language/v6/pkg/walkers/semantics"

	"github.com/yaacov/kubectl-sql/pkg/eval"
)

// Config provides information required filter item list by query.
type Config struct {
	CheckColumnName func(s string) (string, error)
	Query           string

	Prefix1 string
	Prefix2 string
	Item    unstructured.Unstructured
}

// Filter filters items using query.
func (c *Config) Filter(list []unstructured.Unstructured) ([]unstructured.Unstructured, error) {
	var (
		tree *tsl.TSLNode
		err  error
	)

	// If we have a query, prepare the search tree.
	tree, err = tsl.ParseTSL(c.Query)
	if err != nil {
		return nil, err
	}
	defer tree.Free()

	// Check and replace user identifiers if alias exist.
	newTree, err := ident.Walk(tree, c.CheckColumnName)
	if err != nil {
		return nil, err
	}
	defer newTree.Free()

	// Filter items using a query.
	items := []unstructured.Unstructured{}
	for _, item := range list {
		// If we have a query, check item.
		matchingFilter, err := semantics.Walk(newTree, eval.EvalFunctionFactory(item))
		if err != nil {
			continue
		}
		if match, ok := matchingFilter.(bool); ok && match {
			items = append(items, item)
		}
	}

	return items, nil
}

// Filter2 filters items using query and a left side item.
func (c *Config) Filter2(list []unstructured.Unstructured) ([]unstructured.Unstructured, error) {
	var (
		tree *tsl.TSLNode
		err  error
	)

	// If we have a query, prepare the search tree.
	tree, err = tsl.ParseTSL(c.Query)
	if err != nil {
		return nil, err
	}
	defer tree.Free()

	// Check and replace user identifiers if alias exist.
	newTree, err := ident.Walk(tree, c.CheckColumnName)
	if err != nil {
		return nil, err
	}
	defer newTree.Free()

	// Filter items using query.
	items := []unstructured.Unstructured{}
	for _, item := range list {
		// If we have a query, check item.
		matchingFilter, err := semantics.Walk(newTree, eval.JoinEvalFunctionFactory(c.Item, item, c.Prefix1, c.Prefix2))
		if err != nil {
			continue
		}
		if match, ok := matchingFilter.(bool); ok && match {
			items = append(items, item)
		}
	}

	return items, nil
}
