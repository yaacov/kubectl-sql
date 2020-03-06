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

	sqlGetLong = `Uses SQL-like language to filter and display one or many resources.

  Prints information about kubernetes resources filtered using SQL-like query.
If the desired resource type is namespaced you will only see results in your current
namespace unless you pass --all-namespaces. 

Use "%[1]s api-resources" for a complete list of supported resources.`

	sqlGetUsage = `%[1]s sql get <resources> [where "<SQL-like query>"] [flags] [options]`

	sqlGetExample = `  # List all pods in table output format.
  %[1]s sql get pods
	
  # List all replication controllers and services in json output format.
  %[1]s sql get rc,services --output json
  
  # List all pods where name starts with "test-" case insensitive.
  %[1]s sql get pods where "name ilike 'test-%%'"

  # List all pods where the memory request for the first container is lower or equal to 200Mi.
  %[1]s sql --all-namespaces get pods where "spec.containers.1.resources.requests.memory <= 200Mi"`

	errNoContext = fmt.Errorf("no context is currently set, use %q to select a new one", "kubectl config use-context <context>")
	errUsage     = fmt.Errorf("bad command or command usage, use --help flag for help about command usage (kubectl sql [sub-command] --help)")

	defaultAliases = map[string]string{
		"phase": "status.phase",
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
