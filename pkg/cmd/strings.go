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

// SQL command help strings
const (
	// SQLCmdLong is the long description for the sql command
	SQLCmdLong = `
Query Kubernetes resources using SQL-like syntax or subcommands.

This command allows you to use SQL-like queries to fetch and filter Kubernetes resources.
The SQL syntax supports SELECT statements with WHERE clauses to filter resources based on
their attributes.

Examples:
  # Get all pods in the current namespace
  kubectl sql "SELECT name, status.phase FROM pods"

  # Get all pods in all namespaces with specific labels
  kubectl sql "SELECT name, metadata.namespace FROM pods WHERE labels.app = 'nginx'"

  # Show version information
  kubectl sql version
`

	// SQLCmdExample provides examples for the sql command
	SQLCmdExample = `
  # Query pods with phase=Running
  kubectl sql "SELECT name, status.phase FROM pods WHERE status.phase = 'Running'"

  # Get nodes with specific memory capacity
  kubectl sql "SELECT name FROM nodes WHERE status.capacity.memory > 8Gi"

  # Get version information
  kubectl sql version
`

	// ErrorUsageTemplate is the template for usage errors
	ErrorUsageTemplate = "error: %s\nSee 'kubectl sql --help' for usage."

	// VersionCmdShort is the short description for the version command
	VersionCmdShort = "Print the SQL client and server version information"
)
