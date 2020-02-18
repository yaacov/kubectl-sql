// Copyright 2020 Yaacov Zamir <kobi.zamir@gmail.com>
// and other contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Author: 2020 Yaacov Zamir <kobi.zamir@gmail.com>

// Package main.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"k8s.io/client-go/discovery"
)

func main() {
	cli.AppHelpTemplate = helpTemplate()
	cli.VersionPrinter = versionPrinter

	app := &cli.App{
		Name:    "kubesql",
		Version: "v0.0.0",
		Usage:   "uses sql like language to query the Kubernetes cluster manager.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "kubeconfig",
				Value: "",
				Usage: "Path to the kubeconfig file to use for CLI requests.",
			},
			&cli.StringFlag{
				Name:    "namespace",
				Aliases: []string{"n"},
				Value:   "",
				Usage:   "If present, the namespace scope for this CLI request.",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "table",
				Usage:   "Output format, options: table, yaml or json.",
			},
			&cli.BoolFlag{
				Name:    "si-units",
				Aliases: []string{"s"},
				Value:   false,
				Usage:   "Parse values with SI units as numbers, (e.g. '1Ki' will be 1024).",
			},
			&cli.BoolFlag{
				Name:    "all-namespaces",
				Aliases: []string{"A"},
				Value:   false,
				Usage:   "Use all namespace scopes for this CLI request.",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"V"},
				Value:   false,
				Usage:   "Show verbose output",
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "get",
				Usage:  "Display one or many resources.",
				Action: actionsGet,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "resource",
						Aliases: []string{"r"},
						Value:   "",
						Usage:   "If present, the resource `name` to query.",
					},
					&cli.StringFlag{
						Name:    "query",
						Aliases: []string{"q"},
						Value:   "",
						Usage:   "If present, filter results usign SQL like `query`.",
					},
				},
			},
			{
				Name:   "namespace",
				Usage:  "Display current namespace scope.",
				Action: namespacePrinter,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// Print out the namespace message (namespace command).
func namespacePrinter(c *cli.Context) error {
	kubeconfig := getKubeConfig(c)
	namespace, _, err := kubeconfig.Namespace()
	errExit("Failed to get namespace", err)

	fmt.Printf("Current namespace scope: %s\n", namespace)
	return nil
}

// Print out the version message ( -v --version flags ).
func versionPrinter(c *cli.Context) {
	kubeconfig := getKubeConfig(c)

	config, err := kubeconfig.ClientConfig()
	errExit("Failed to load client conifg", err)

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	errExit("Failed to create discovery client", err)

	serverVersion, err := discoveryClient.ServerVersion()

	if err != nil {
		fmt.Printf(versionTemplate(), c.App.Version)
	} else {
		fmt.Printf(fullVersionTemplate(), c.App.Version, serverVersion)
	}
}

// Print error and exit.
func errExit(msg string, err error) {
	if err != nil {
		log.Fatalf("%s: %#v", msg, err)
	}
}
