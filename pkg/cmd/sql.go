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
	"sort"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
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
		Use:              "sql [command] [flags] [options]",
		Short:            sqlCmdShort,
		Long:             sqlCmdLong,
		Example:          sqlCmdExample,
		TraverseChildren: true,
		RunE: func(c *cobra.Command, args []string) error {
			return fmt.Errorf(errUsageTemplate, "missing sub command")
		},
	}

	cmdGet := &cobra.Command{
		Use:              "get <resource[,...]> [where \"<SQL-like query>\"] [flags] [options]",
		Short:            sqlGetShort,
		Long:             sqlGetLong,
		Example:          sqlGetExample,
		TraverseChildren: true,
		SilenceUsage:     true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}

			if err := o.CompleteGet(c, args); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			config, err := o.rawConfig.ClientConfig()
			if err != nil {
				return err
			}

			if err := o.Get(config); err != nil {
				return err
			}

			return nil
		},
	}

	cmdJoin := &cobra.Command{
		Use:              "join <resource,resource> on \"<SQL-like query>\" [where \"<SQL-like query>\"] [flags] [options]",
		Short:            sqlJoinShort,
		Long:             sqlJoinLong,
		Example:          sqlJoinExample,
		TraverseChildren: true,
		SilenceUsage:     true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}

			if err := o.CompleteJoin(c, args); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			config, err := o.rawConfig.ClientConfig()
			if err != nil {
				return err
			}

			if err := o.Join(config); err != nil {
				return err
			}

			return nil
		},
	}

	cmdAliases := &cobra.Command{
		Use:          "aliases [flags] [options]",
		Short:        "Display a list of currently used aliases",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}

			// Read SQL plugin specific configurations.
			if err := o.readConfigFile(o.requestedSQLConfigPath); err != nil {
				return err
			}

			hardCodedAliases := map[string]string{
				"name":      "resource name",
				"namespace": "resource namespace",
				"created":   "resource creation time",
				"deleted":   "resource delition time",
			}

			fmt.Fprintf(o.Out, "%-12s\t%s\n", "ALIAS", "PATH")

			// Print hard coded alias table.
			for k, v := range hardCodedAliases {
				fmt.Fprintf(o.Out, "%-12s\t%s\n", k, v)
			}

			// Print alias table, sorted by keys.
			keys := sortedKeys(o.defaultAliases)
			for _, k := range keys {
				fmt.Fprintf(o.Out, "%-12s\t%s\n", k, o.defaultAliases[k])
			}

			return nil
		},
	}

	cmdVersion := &cobra.Command{
		Use:          "version [flags]",
		Short:        "Print the SQL client and server version information",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			config, err := o.rawConfig.ClientConfig()
			if err != nil {
				return err
			}

			if err := o.Version(config); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.AddCommand(cmdGet, cmdJoin, cmdVersion, cmdAliases)

	cmd.Flags().BoolVarP(&o.allNamespaces, "all-namespaces", "A", o.allNamespaces,
		"If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmd.Flags().StringVarP(&o.requestedSQLConfigPath, "kubectl-sql", "q", o.defaultSQLConfigPath,
		"Path to the kubectl-sql.json file to use for kubectl-sql requests.")
	cmd.Flags().StringVarP(&o.outputFormat, "output", "o", o.outputFormat,
		"Output format. One of: json|yaml|table|name")

	o.configFlags.AddFlags(cmd.Flags())

	cmdGet.Flags().AddFlagSet(cmd.Flags())
	cmdJoin.Flags().AddFlagSet(cmd.Flags())
	cmdAliases.Flags().AddFlagSet(cmd.Flags())
	cmdVersion.Flags().AddFlagSet(cmd.Flags())

	return cmd
}

// Complete sets all information required for updating the current context
func (o *SQLOptions) Complete(cmd *cobra.Command, args []string) error {
	var err error
	o.args = args

	o.rawConfig = o.configFlags.ToRawKubeConfigLoader()
	if o.namespace, _, err = o.rawConfig.Namespace(); err != nil {
		return err
	}

	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *SQLOptions) Validate() error {
	formatOptions := map[string]bool{"table": true, "json": true, "yaml": true, "name": true}

	if _, ok := formatOptions[o.outputFormat]; !ok {
		return fmt.Errorf("output format must be one of: json|yaml|table|name")
	}

	return nil
}

// Get sorted map keys
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
