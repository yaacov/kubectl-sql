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

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"

	"github.com/yaacov/kubectl-sql/pkg/client"
)

// NewVersionCmd creates a command that displays version information
func NewVersionCmd(streams *genericIOStreams) *cobra.Command {
	o := NewSQLOptions(genericclioptions.IOStreams{
		In:     streams.In,
		Out:    streams.Out,
		ErrOut: streams.ErrOut,
	})

	cmd := &cobra.Command{
		Use:          "version",
		Short:        VersionCmdShort,
		SilenceUsage: true,
		RunE: func(versionCmd *cobra.Command, args []string) error {
			config, err := client.NewFromCLIArgs(versionCmd, args, o.configFlags)
			if err != nil {
				return err
			}

			return ShowVersion(&streams.Out, config)
		},
	}

	// Add config flags to the version command
	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}

// ShowVersion prints the client and server version information
func ShowVersion(w *io.Writer, config *client.Config) error {
	serverVersionStr := "unknown"

	// Get server version if possible
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config.Config)
	if err == nil {
		serverVersion, err := discoveryClient.ServerVersion()
		if err == nil {
			serverVersionStr = fmt.Sprintf("%v", serverVersion)
		}
	}

	// Print version information
	fmt.Fprintf(*w, "Client version: %s\n", clientVersion)
	fmt.Fprintf(*w, "Server version: %s\n", serverVersionStr)
	fmt.Fprintf(*w, "Current namespace: %s\n", config.Namespace)

	return nil
}
