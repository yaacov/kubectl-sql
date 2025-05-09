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

package execute

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-sql/pkg/client"
	"github.com/yaacov/kubectl-sql/pkg/query"
	"github.com/yaacov/kubectl-sql/pkg/resourcefields"
)

// Query parses and executes an SQL query against Kubernetes resources
func Query(c *cobra.Command, args []string, configFlags *genericclioptions.ConfigFlags, debugLevel int) ([]map[string]interface{}, *query.QueryOptions, error) {
	ctx := c.Context()

	// Create client configuration, passing the existing config flags
	config, err := client.NewFromCLIArgs(c, args, configFlags)
	if err != nil {
		return nil, nil, err
	}

	// Parse the SQL query
	queryOptions, err := query.ParseQueryString(args[0])
	if err != nil {
		return nil, nil, err
	}

	// Set the debug level in queryOptions
	queryOptions.DebugLevel = debugLevel

	// Print queryOptions in JSON format if debug is enabled
	if debugLevel > 0 {
		queryOptionsJSON, err := json.MarshalIndent(queryOptions, "", "  ")
		if err == nil {
			fmt.Fprintf(c.OutOrStderr(), "DEBUG: Query options after parsing:\n%s\n", string(queryOptionsJSON))
		} else {
			fmt.Fprintf(c.OutOrStderr(), "DEBUG: Failed to marshal query options: %v\n", err)
		}
	}

	// Validate query has FROM clause
	if !queryOptions.HasFrom {
		return nil, nil, fmt.Errorf("query must include a FROM clause specifying the resource type")
	}

	// If the select clause "*" set default fields
	if len(queryOptions.Select) == 1 && queryOptions.Select[0].Field == ".*" {
		queryOptions.Select = resourcefields.GetDefaultSelectFields(queryOptions.From.Name)
	}

	// Get resource name and namespace from the query
	resourceName := queryOptions.From.Name
	resourceNamespace := queryOptions.From.Namespace
	allNamespaces := queryOptions.From.AllNamespaces

	// Get the resource list from the Kubernetes API
	resourceList, err := config.ListAsMap(ctx, resourceName, resourceNamespace, allNamespaces)
	if err != nil {
		return nil, nil, err
	}

	// Execute query
	result, err := query.ApplyQuery(resourceList, queryOptions)
	if err != nil {
		return nil, nil, err
	}

	return result, queryOptions, nil
}
