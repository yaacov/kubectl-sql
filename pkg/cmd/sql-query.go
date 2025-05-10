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
	"io"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-sql/pkg/cmd/execute"
)

// genericIOStreams wraps the genericclioptions.IOStreams for easier testing
type genericIOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

// SQLOptions provides information required to run kubectl-sql
type SQLOptions struct {
	configFlags  *genericclioptions.ConfigFlags
	streams      *genericIOStreams
	outputFormat string
	noHeaders    bool
	debugLevel   int
}

// NewSQLOptions returns initialized SQLOptions
func NewSQLOptions(streams genericclioptions.IOStreams) *SQLOptions {
	return &SQLOptions{
		configFlags:  genericclioptions.NewConfigFlags(true),
		streams:      &genericIOStreams{In: streams.In, Out: streams.Out, ErrOut: streams.ErrOut},
		outputFormat: "table", // Default output format
		debugLevel:   0,       // Default debug level (off)
	}
}

// NewCmdSQL provides a cobra command wrapping SQLOptions
func NewCmdSQL(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewSQLOptions(streams)

	cmd := &cobra.Command{
		Use:              "[query|command] [flags] [options]",
		Short:            "Query Kubernetes resources using SQL-like syntax or subcommands",
		Long:             SQLCmdLong,
		Example:          SQLCmdExample,
		TraverseChildren: true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("missing query or sub command, see 'kubectl sql --help'")
			}

			// Execute SQL query if the first argument looks like SQL
			if strings.HasPrefix(strings.ToUpper(args[0]), "SELECT") {
				result, queryOptions, err := execute.Query(c, args, o.configFlags, o.debugLevel)
				if err != nil {
					return err
				}

				// Print results according to output format
				return PrintResults(o, result, queryOptions)
			}

			// Not an SQL query and not a recognized subcommand
			return fmt.Errorf("unrecognized command, see 'kubectl sql --help'")
		},
	}

	// Add subcommands
	cmd.AddCommand(NewVersionCmd(o.streams))

	// Add flags
	cmd.Flags().StringVarP(&o.outputFormat, "output", "o", o.outputFormat,
		"Output format. One of: json|yaml|table|name")
	cmd.Flags().BoolVarP(&o.noHeaders, "no-headers", "H", false,
		"When using the table output format, don't print headers (column titles)")
	cmd.Flags().IntVar(&o.debugLevel, "debug", o.debugLevel,
		"Debug level: 0=off, 1=info, 2=verbose (default 0)")

	// Add flags for generic options
	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}
