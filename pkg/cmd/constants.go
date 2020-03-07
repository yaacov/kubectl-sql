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
	"fmt"

	"github.com/yaacov/kubectl-sql/pkg/printers"
)

var (
	clientVersion = "GIT-master"

	// sql command.
	sqlCmdShort = "Uses SQL-like language to filter and display one or many resources"
	sqlCmdLong  = `Uses SQL-like language to filter and display one or many resources.

  kubectl sql prints information about kubernetes resources filtered using SQL-like query`

	sqlCmdExample = `  # Print client version.
  kubectl sql version

  # Print this help message.
  kubectl sql help

  # List all pods where name starts with "test-" case insensitive.
  kubectl sql get pods where "name ilike 'test-%%'"

  # List all virtual machine instanaces and pods joined on vim is owner of pod.
  kubectl sql join vmis,pods on "vmis.metadata.uid = pods.metadata.ownerReferences.1.uid"`

	// sql get command.
	sqlGetShort = "Uses SQL-like language to filter and display one or many resources"
	sqlGetLong  = `Uses SQL-like language to filter and display one or many resources.

  kubectl sql get prints information about kubernetes resources filtered using SQL-like query.
If the desired resource type is namespaced you will only see results in your current
namespace unless you pass --all-namespaces`

	sqlGetExample = `  # List all pods in table output format.
  kubectl sql get pods
	
  # List all replication controllers and services in json output format.
  kubectl sql get rc,services --output json
  
  # List all pods where name starts with "test-" case insensitive.
  kubectl sql get pods where "name ilike 'test-%%'"

  # List all pods where the memory request for the first container is lower or equal to 200Mi.
  kubectl sql --all-namespaces get pods where "spec.containers.1.resources.requests.memory <= 200Mi"`

	// sql get command.
	sqlJoinShort = "Uses SQL-like language to join two resources"
	sqlJoinLong  = `Uses SQL-like language to join two resources.

  kubectl sql join prints information about kubernetes resources joined using SQL-like query.
If the desired resource type is namespaced you will only see results in your current
namespace unless you pass --all-namespaces`

	sqlJoinExample = `  # List all virtual machine instanaces and pods joined on vim is owner of pod.
  kubectl sql join vmis,pods on "vmis.metadata.uid = pods.metadata.ownerReferences.1.uid"

  # List all virtual machine instanaces and pods joined on vim is owner of pod for vmis with name matching 'test' regexp.
  kubectl sql join vmis,pods on "vmis.metadata.uid = pods.metadata.ownerReferences.1.uid" where "name ~= 'test'" -A

  # Join deployment sets with pods using the uid aliases.
  kubectl sql join ds,pods on "ds.uid = pods.owner.uid"

  # Display non running pods by nodes for all namespaces.
  kubectl sql join nodes,pods on "nodes.status.addresses.1.address = pods.status.hostIP and not pods.phase ~= 'Running'" -A `

	// sql version command
	sqlVersionShort = "Print the SQL client and server version information"
	sqlVersionLong  = "Print the SQL client and server version information."

	sqlVersionExample = `# Print the SQL client and server versions for the current context
  kubectl sql version"`

	// Errors.
	errNoContext     = fmt.Errorf("no context is currently set, use %q to select a new one", "kubectl config use-context <context>")
	errUsageTemplate = "bad command or command usage, %s"

	// Defaults.
	defaultAliases = map[string]string{
		"phase":     "status.phase",
		"uid":       "metadata.uid",
		"owner.uid": "metadata.ownerReferences.1.uid",
	}
	defaultTableFields = printers.TableFieldsMap{
		"other": {
			{
				Title: "NAMESPACE",
				Name:  "namespace",
			},
			{
				Title: "NAME",
				Name:  "name",
			},
			{
				Title: "PHASE",
				Name:  "status.phase",
			},
			{
				Title: "CREATION_TIME(RFC3339)",
				Name:  "created",
			},
		},
	}
)
