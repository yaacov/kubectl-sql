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
	"strings"

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

	// Initialize default configuration
	initializeDefaults(options)

	return options
}

// NewCmdSQL provides a cobra command wrapping SQLOptions
func NewCmdSQL(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewSQLOptions(streams)

	cmd := &cobra.Command{
		Use:              "sql <query> [flags] [options]",
		Short:            "Query Kubernetes resources using SQL-like syntax",
		Long:             sqlCmdLong,
		Example:          sqlCmdExample,
		TraverseChildren: true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf(errUsageTemplate, "missing SQL query")
			}

			// All arguments should be treated as a single SQL query
			query := strings.Join(args, " ")

			if err := o.Complete(c, args); err != nil {
				return err
			}

			if err := o.CompleteSQL(query); err != nil {
				return err
			}

			if err := o.Validate(); err != nil {
				return err
			}

			config, err := o.rawConfig.ClientConfig()
			if err != nil {
				return err
			}

			// Execute query based on number of resources
			if len(o.requestedResources) >= 1 {
				return o.Get(config)
			} else {
				return fmt.Errorf("invalid number of resources in query")
			}
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

	cmd.AddCommand(cmdVersion)

	cmd.Flags().StringVarP(&o.outputFormat, "output", "o", o.outputFormat,
		"Output format. One of: json|yaml|table|name")
	cmd.Flags().BoolVarP(&o.noHeaders, "no-headers", "H", false,
		"When using the table output format, don't print headers (column titles)")

	o.configFlags.AddFlags(cmd.Flags())

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
