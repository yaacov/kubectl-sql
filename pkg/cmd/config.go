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
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/yaacov/kubectl-sql/pkg/printers"
)

// SQLOptions provides information required to update
// the current context on a user's KUBECONFIG
type SQLOptions struct {
	configFlags *genericclioptions.ConfigFlags

	rawConfig clientcmd.ClientConfig
	args      []string

	defaultAliases     map[string]string
	defaultTableFields printers.TableFieldsMap
	orderByFields      []printers.OrderByField
	limit              int

	namespace          string
	requestedResources []string
	requestedQuery     string

	outputFormat string
	noHeaders    bool

	genericclioptions.IOStreams
}

// NewSQLOptions provides an instance of SQLOptions with default values initialized
func initializeDefaults(o *SQLOptions) {
	o.defaultAliases = defaultAliases
	o.defaultTableFields = defaultTableFields
	o.orderByFields = []printers.OrderByField{}
	o.limit = 0 // Default to no limit
}
