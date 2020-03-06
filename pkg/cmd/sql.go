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
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	clientVersion = "GIT-master"

	sqlGetLong = `Uses SQL-like language to filter and display one or many resources.

  Prints information about kubernetes resources filtered using SQL-like query.
If the desired resource type is namespaced you will only see results in your current
namespace unless you pass --all-namespaces. 

Use "%[1]s api-resources" for a complete list of supported resources.`

	sqlGetUsage = `%[1]s sql get <resources> [where "<SQL-like query>"] [flags] [options]`

	sqlGetExample = `  # List all pods in ps output format.
  %[1]s sql get pods

  # List deployments in JSON output format, in the "v1" version of the "apps" API group:
  %[1]s sql get deployments.v1.apps -o json
	
  # List all replication controllers and services together in ps output format.
  %[1]s sql get rc,services`

	errNoContext = fmt.Errorf("no context is currently set, use %q to select a new one", "kubectl config use-context <context>")
	errUsage     = fmt.Errorf("Use: get <resources> [where <SQL-like query>]")
)

// NewSQLOptions provides an instance of SQLOptions with default values
func NewSQLOptions(streams genericclioptions.IOStreams) *SQLOptions {
	options := &SQLOptions{
		configFlags:  genericclioptions.NewConfigFlags(true),
		IOStreams:    streams,
		outputFormat: "table",
	}

	// Look for a default kubectl-sql.json config file.
	if home, err := os.UserHomeDir(); err == nil {
		options.defaultSQLConfigPath = fmt.Sprintf("%s/.kube/kubectl-sql.json", home)
	}

	return options
}

// NewCmdSQL provides a cobra command wrapping SQLOptions
func NewCmdSQL(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewSQLOptions(streams)

	cmd := &cobra.Command{
		Long:         fmt.Sprintf(sqlGetLong, "kubectl"),
		Use:          fmt.Sprintf(sqlGetUsage, "kubectl"),
		Example:      fmt.Sprintf(sqlGetExample, "kubectl"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&o.allNamespaces, "all-namespaces", "A", o.allNamespaces,
		"If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmd.Flags().StringVarP(&o.requestedSQLConfigPath, "kubectl-sql", "q", o.defaultSQLConfigPath,
		"Path to the kubectl-sql.json file to use for kubectl-sql requests.")
	cmd.Flags().StringVarP(&o.outputFormat, "output", "o", o.outputFormat,
		"Output format. One of: json|yaml|table|name")

	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}

// Complete sets all information required for updating the current context
func (o *SQLOptions) Complete(cmd *cobra.Command, args []string) error {
	var err error
	subCommandsArgs := map[int]string{1: "version", 2: "get", 4: "get"}

	o.args = args

	if len(o.args) == 0 {
		return errUsage
	}

	o.rawConfig = o.configFlags.ToRawKubeConfigLoader()
	if o.namespace, _, err = o.rawConfig.Namespace(); err != nil {
		return err
	}

	// Read SQL plugin specific configurations.
	if err = o.readConfigFile(o.requestedSQLConfigPath); err != nil {
		return err
	}

	// Parse SQL sub command.
	o.subCommand = strings.ToLower(o.args[0])
	if o.subCommand != subCommandsArgs[len(o.args)] {
		return errUsage
	}

	if o.subCommand == "get" {
		o.requestedResources = strings.Split(o.args[1], ",")

		// Look for where
		if len(o.args) == 4 {
			if strings.ToLower(o.args[2]) != "where" {
				return errUsage
			}

			o.requestedQuery = o.args[3]
		}
	}

	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *SQLOptions) Validate() error {
	formatOptions := map[string]bool{"table": true, "json": true, "yaml": true, "name": true}

	if _, ok := formatOptions[o.outputFormat]; !ok {
		return fmt.Errorf("output format must be one of: json|yaml|table|name")
	}

	if o.requestedSQLConfigPath != o.defaultSQLConfigPath && !fileExists(o.requestedSQLConfigPath) {
		return fmt.Errorf("can't open '%s', file may be missing", o.requestedSQLConfigPath)
	}

	return nil
}

// Run the SQL sub command.
func (o *SQLOptions) Run() error {
	config, err := o.rawConfig.ClientConfig()
	if err != nil {
		return err
	}

	// Print plugin version (sub-command = "version").
	if o.subCommand == "version" {
		return o.Version(config)
	}

	// Print resources lists.
	if o.subCommand == "get" {
		return o.Get(config)
	}

	return nil
}
