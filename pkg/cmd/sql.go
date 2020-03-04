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

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	clientVersion = "v0.2.0"

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
)

// SQLOptions provides information required to update
// the current context on a user's KUBECONFIG
type SQLOptions struct {
	configFlags *genericclioptions.ConfigFlags

	rawConfig              clientcmd.ClientConfig
	namespace              string
	allNamespaces          bool
	defaultSQLConfigPath   string
	requestedSQLConfigPath string
	outputFormat           string
	args                   []string

	defaultAliases     map[string]string
	defaultTableFields tableFieldsMap

	requestedResources []string
	requestedQuery     string

	version bool

	genericclioptions.IOStreams
}

// NewSQLOptions provides an instance of SQLOptions with default values
func NewSQLOptions(streams genericclioptions.IOStreams) *SQLOptions {
	options := &SQLOptions{
		configFlags:  genericclioptions.NewConfigFlags(true),
		IOStreams:    streams,
		outputFormat: "table",
	}

	// Look for a default kubesql.json config file.
	if home, err := os.UserHomeDir(); err == nil {
		options.defaultSQLConfigPath = fmt.Sprintf("%s/.kube/kubesql.json", home)
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
	cmd.Flags().StringVarP(&o.requestedSQLConfigPath, "kubesql", "q", o.defaultSQLConfigPath,
		"Path to the kubesql.json file to use for kubesql requests.")
	cmd.Flags().StringVarP(&o.outputFormat, "output", "o", o.outputFormat,
		"Output format. One of: json|yaml|table|name")

	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}

// Complete sets all information required for updating the current context
func (o *SQLOptions) Complete(cmd *cobra.Command, args []string) error {
	var err error
	queryType := map[int]string{1: "version", 2: "get", 4: "get-where"}

	o.args = args
	o.rawConfig = o.configFlags.ToRawKubeConfigLoader()
	if o.namespace, _, err = o.rawConfig.Namespace(); err != nil {
		return err
	}

	// Read SQL plugin specific configurations.
	if err = o.readConfigFile(o.requestedSQLConfigPath); err != nil {
		return err
	}

	// Parse user request.
	switch queryType[len(o.args)] {
	case "version":
		o.version = true
	case "get":
		o.requestedResources = strings.Split(o.args[1], ",")
	case "get-where":
		o.requestedResources = strings.Split(o.args[1], ",")
		o.requestedQuery = o.args[3]
	}

	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *SQLOptions) Validate() error {
	formatOptions := map[string]bool{"table": true, "json": true, "yaml": true, "name": true}
	queryType := map[int]string{1: "version", 2: "get", 4: "get-where"}

	if _, ok := queryType[len(o.args)]; !ok {
		return fmt.Errorf("Use: get <resources> [where <SQL-like query>]")
	}
	switch queryType[len(o.args)] {
	case "version":
		if strings.ToLower(o.args[0]) != "version" {
			return fmt.Errorf("Use: get <resources> [where <SQL-like query>]")
		}
	case "get":
		if strings.ToLower(o.args[0]) != "get" {
			return fmt.Errorf("Use: get <resources> [where <SQL-like query>]")
		}
	case "get-where":
		if strings.ToLower(o.args[0]) != "get" || strings.ToLower(o.args[2]) != "where" {
			return fmt.Errorf("Use: get <resources> [where <SQL-like query>]")
		}
	}

	if _, ok := formatOptions[o.outputFormat]; !ok {
		return fmt.Errorf("output format must be one of: json|yaml|table|name")
	}

	if o.requestedSQLConfigPath != o.defaultSQLConfigPath && !fileExists(o.requestedSQLConfigPath) {
		return fmt.Errorf("can't open '%s', file may be missing", o.requestedSQLConfigPath)
	}

	return nil
}

// Run lists all available namespaces on a user's KUBECONFIG or updates the
// current context based on a provided namespace.
func (o *SQLOptions) Run() error {
	config, err := o.rawConfig.ClientConfig()
	if err != nil {
		return err
	}

	// Print plugin version (sub-command = "version").
	if o.version {
		return o.Version(config)
	}

	// Print resource list.
	list, err := o.List(config, o.requestedResources[0], o.requestedQuery)
	if err != nil {
		return err
	}

	err = o.Printer(list)
	if err != nil {
		return err
	}

	return nil
}

// Version prints the plugin version.
func (o *SQLOptions) Version(config *rest.Config) error {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return err
	}

	serverVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		return err
	}

	fmt.Fprintf(o.Out, "Client version: %v\n", clientVersion)
	fmt.Fprintf(o.Out, "Server version: %v\n", serverVersion)
	fmt.Fprintf(o.Out, "Current namespace: %s\n", o.namespace)

	return nil
}
