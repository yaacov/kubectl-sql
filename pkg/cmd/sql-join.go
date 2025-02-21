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
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"

	"github.com/yaacov/kubectl-sql/pkg/client"
	"github.com/yaacov/kubectl-sql/pkg/filter"
)

// CompleteJoin sets all information required for updating the current context for join sub command.
func (o *SQLOptions) CompleteJoin(cmd *cobra.Command, args []string) error {
	var err error
	o.args = args

	if len(o.args) != 5 && len(o.args) != 3 {
		return fmt.Errorf(errUsageTemplate, "bad number of arguments")
	}

	// Read SQL plugin specific configurations.
	if err = o.readConfigFile(o.requestedSQLConfigPath); err != nil {
		return err
	}

	// join <resource,resource> on <query> where <query>
	o.requestedResources = strings.Split(o.args[0], ",")
	if len(o.requestedResources) != 2 {
		return fmt.Errorf(errUsageTemplate, "join command takes exectly two resources")
	}

	// Look for "on"
	if strings.ToLower(o.args[1]) != "on" {
		return fmt.Errorf(errUsageTemplate, "missing \"on\" argument")
	}

	// Look for "where"
	if len(o.args) == 5 {
		if strings.ToLower(o.args[3]) != "where" {
			return fmt.Errorf(errUsageTemplate, "missing \"where\" argument")
		}
		o.requestedQuery = o.args[4]
	}

	o.requestedOnQuery = o.args[2]

	return nil
}

// Join two resource list.
func (o *SQLOptions) Join(config *rest.Config) error {
	ctx := context.Background()
	var err error
	var filteredList []unstructured.Unstructured

	c := client.Config{
		Config:        config,
		Namespace:     o.namespace,
		AllNamespaces: o.allNamespaces,
	}

	f := filter.Config{
		CheckColumnName: o.checkColumnName,
		Query:           o.requestedQuery,
	}

	// Get the primary resources lists.
	list1, err := c.List(ctx, o.requestedResources[0])
	if err != nil {
		return err
	}

	// Get the joined resources lists.
	list2, err := c.List(ctx, o.requestedResources[1])
	if err != nil {
		return err
	}

	// Filter primary list if needed.
	if len(o.requestedQuery) > 0 {
		// Filter primary items by query.
		filteredList, err = f.Filter(list1)
		if err != nil {
			return err
		}
	} else {
		filteredList = list1
	}

	for _, r := range filteredList {
		// Print one item.
		err := o.Printer([]unstructured.Unstructured{r})
		if err != nil {
			return err
		}

		// Print separator.
		fmt.Fprintf(o.Out, "\n")

		// Print joined items.
		if err := o.printJoinedResources(r, list2); err != nil {
			return err
		}

		// Print separator.
		fmt.Fprintf(o.Out, "\n\n\n")
	}

	return nil
}

// printJoinedResources prints joined resource list.
func (o *SQLOptions) printJoinedResources(item unstructured.Unstructured, list2 []unstructured.Unstructured) error {
	f := filter.Config{
		CheckColumnName: o.checkColumnName2,
		Query:           o.requestedOnQuery,

		Prefix1: o.requestedResources[0],
		Prefix2: o.requestedResources[1],
		Item:    item,
	}

	// Filter joined items by primary item and "on" query.
	filteredList, err := f.Filter2(list2)
	if err != nil {
		return err
	}

	err = o.Printer(filteredList)
	if err != nil {
		return err
	}

	return nil
}

// checkColumnName2 checks if a coulumn name has an alias using prefixes.
func (o *SQLOptions) checkColumnName2(s string) (string, error) {
	var (
		prefix1 = o.requestedResources[0]
		prefix2 = o.requestedResources[1]
	)

	if strings.HasPrefix(s, prefix1+".") {
		if v, ok := o.defaultAliases[strings.TrimPrefix(s, prefix1+".")]; ok {
			return prefix1 + "." + v, nil
		}
	} else if strings.HasPrefix(s, prefix2+".") {
		if v, ok := o.defaultAliases[strings.TrimPrefix(s, prefix2+".")]; ok {
			return prefix2 + "." + v, nil
		}
	} else if v, ok := o.defaultAliases[s]; ok {
		return v, nil
	}

	// If not found in alias table, return the column name unchanged.
	return s, nil
}
