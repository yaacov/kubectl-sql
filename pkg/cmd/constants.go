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
	"github.com/yaacov/kubectl-sql/pkg/printers"
)

var (
	clientVersion = "GIT-master"

	// sql command.
	sqlCmdLong = `Uses SQL-like language to filter and display Kubernetes resources.

  kubectl sql prints information about kubernetes resources filtered using SQL-like query`

	sqlCmdExample = `  # Print client version.
  kubectl sql version

  # List all pods where name starts with "test-" case insensitive.
  kubectl sql "select * from pods where name ilike 'test-%%'"
  
  # List first 5 pods ordered by creation time in descending order (newest first).
  kubectl sql "select * from pods order by created desc limit 5"

  # Print this help message.
  kubectl sql help`

	// Errors.
	errUsageTemplate = "bad command or command usage, %s"

	// Defaults.
	defaultAliases = map[string]string{
		"name":      "metadata.name",
		"namespace": "metadata.namespace",
		"created":   "metadata.creationTimestamp",
		"phase":     "status.phase",
		"uid":       "metadata.uid",
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
