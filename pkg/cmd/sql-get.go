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
	"github.com/yaacov/kubectl-sql/pkg/printers"
)

// CompleteGet sets all information required for updating the current context for get sub command.
func (o *SQLOptions) CompleteGet(cmd *cobra.Command, args []string) error {
	var err error
	o.args = args

	if len(o.args) != 1 && len(o.args) != 3 {
		return fmt.Errorf(errUsageTemplate, "bad number of arguments")
	}

	// Read SQL plugin specific configurations.
	if err = o.readConfigFile(o.requestedSQLConfigPath); err != nil {
		return err
	}

	// get <resource list> [where <query>]
	o.requestedResources = strings.Split(o.args[0], ",")

	// Look for "where"
	if len(o.args) == 3 {
		if strings.ToLower(o.args[1]) != "where" {
			return fmt.Errorf(errUsageTemplate, "missing \"where\" argument")
		}

		o.requestedQuery = o.args[2]
	}

	return nil
}

// Get the resource list.
func (o *SQLOptions) Get(config *rest.Config) error {
	c := client.Config{
		Config:        config,
		Namespace:     o.namespace,
		AllNamespaces: o.allNamespaces,
	}

	if len(o.requestedQuery) > 0 {
		return o.printFilteredResources(c)
	}

	return o.printResources(c)
}

// printResources prints resources lists.
func (o *SQLOptions) printResources(c client.Config) error {
	ctx := context.Background()
	for _, r := range o.requestedResources {
		list, err := c.List(ctx, r)
		if err != nil {
			return err
		}

		err = o.Printer(list)
		if err != nil {
			return err
		}
	}

	return nil
}

// printFilteredResources prints filtered resource list.
func (o *SQLOptions) printFilteredResources(c client.Config) error {
	ctx := context.Background()
	f := filter.Config{
		CheckColumnName: o.checkColumnName,
		Query:           o.requestedQuery,
	}

	// Print resources lists.
	for _, r := range o.requestedResources {
		list, err := c.List(ctx, r)
		if err != nil {
			return err
		}

		// Filter items by query.
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

// checkColumnName checks if a column name has an alias.
func (o *SQLOptions) checkColumnName(s string) (string, error) {
	// Check for aliases.
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
		TableFields:   o.defaultTableFields,
		OrderByFields: o.orderByFields,
		Limit:         o.limit,
		Out:           o.Out,
		ErrOut:        o.ErrOut,
		NoHeaders:     o.noHeaders,
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
		err := p.Table(items)
		if err != nil {
			return err
		}
	}

	return nil
}
