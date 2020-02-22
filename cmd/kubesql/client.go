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
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/urfave/cli/v2"
)

// Check if a string in slice of strings.
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Get kubeconfig
func getKubeConfig(c *cli.Context) clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	// User overides default config file.
	if len(c.String("kubeconfig")) > 0 {
		loadingRules.ExplicitPath = c.String("kubeconfig")
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules,
		configOverrides)

	return kubeConfig
}

// Look for a resource matching request resource name.
func getResource(config *rest.Config, resourceName string) (v1.APIResource, *v1.APIResourceList) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	errExit("Failed to create discovery client", err)

	resources, err := discoveryClient.ServerPreferredResources()
	errExit("Failed to create discovery client", err)

	// Search for a matching resource
	resource := v1.APIResource{}
	resourceList := &v1.APIResourceList{}
	for _, rl := range resources {
		for _, r := range rl.APIResources {
			names := append(r.ShortNames, r.Name)
			if stringInSlice(resourceName, names) {
				resource = r
				resourceList = rl
			}
		}

		if len(resource.Name) > 0 {
			break
		}
	}

	if len(resource.Name) == 0 {
		errExit("Failed to find resource", fmt.Errorf("missing resource in server"))
	}

	return resource, resourceList
}

// Get resource group and version.
func getGroupVersion(resourceList *v1.APIResourceList) (string, string) {
	group := ""
	version := ""
	resourceGroupSplit := strings.Split(resourceList.GroupVersion, "/")
	if len(resourceGroupSplit) == 2 {
		group = resourceGroupSplit[0]
		version = resourceGroupSplit[1]
	} else {
		version = resourceGroupSplit[0]
	}

	return group, version
}

// Get interactive namespace usign kubeconfig and flags.
func getNamespace(c *cli.Context, kubeconfig clientcmd.ClientConfig) string {
	if c.Bool("all-namespaces") {
		return ""
	}

	if namespace := c.String("namespace"); len(namespace) > 0 {
		return namespace
	}

	namespace, _, err := kubeconfig.Namespace()
	errExit("Failed to get namespace", err)

	return namespace
}
