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

package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
)

// Config provides information required to query the kubernetes server.
type Config struct {
	Config    *rest.Config
	Namespace string
}

// Complete sets all information required for updating the current context
func NewFromCLIArgs(cmd *cobra.Command, args []string, configFlags *genericclioptions.ConfigFlags) (*Config, error) {
	var err error
	c := &Config{}

	// If no config flags provided, create new ones
	if configFlags == nil {
		configFlags = genericclioptions.NewConfigFlags(true)
		configFlags.AddFlags(cmd.Flags())
	}

	// Read kubeconfig and set the default namespace if not specified
	rawConfig := configFlags.ToRawKubeConfigLoader()
	if c.Namespace, _, err = rawConfig.Namespace(); err != nil {
		return c, err
	}

	c.Config, err = rawConfig.ClientConfig()
	if err != nil {
		return c, err
	}

	return c, nil
}

// List resources by resource name.
func (c *Config) List(ctx context.Context, resourceName, resourceNamespace string, allNamespaces bool) ([]unstructured.Unstructured, error) {
	var err error
	var list *unstructured.UnstructuredList

	// Check for empty resource name
	if len(resourceName) == 0 {
		return nil, fmt.Errorf("resource name is empty")
	}

	// Check for empty resource namespace
	if len(resourceNamespace) == 0 {
		resourceNamespace = c.Namespace
	}

	resource, group, version, err := c.getResourceGroupVersion(resourceName)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(c.Config)
	if err != nil {
		return nil, err
	}

	// Get all resource objects.
	res := dynamicClient.Resource(schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource.Name,
	})

	// Check for namespace
	if !allNamespaces && len(resourceNamespace) > 0 && resource.Namespaced {
		list, err = res.Namespace(resourceNamespace).List(ctx, v1.ListOptions{})
	} else {
		list, err = res.List(ctx, v1.ListOptions{})
	}

	if err != nil {
		return nil, err
	}

	return list.Items, err
}

// ListAsMap resources by resource name and returns them as a slice of maps.
func (c *Config) ListAsMap(ctx context.Context, resourceName, resourceNamespace string, allNamespaces bool) ([]map[string]interface{}, error) {
	items, err := c.List(ctx, resourceName, resourceNamespace, allNamespaces)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(items))
	for i, item := range items {
		result[i] = item.Object
	}

	return result, nil
}

// Look for a resource matching request resource name.
func (c *Config) getResourceGroupVersion(resourceName string) (v1.APIResource, string, string, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(c.Config)
	if err != nil {
		return v1.APIResource{}, "", "", err
	}

	resources, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return v1.APIResource{}, "", "", err
	}

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
		return v1.APIResource{}, "", "", fmt.Errorf("failed to find resource")
	}

	group, version := getGroupVersion(resourceList)
	return resource, group, version, nil
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

// Check if a string in slice of strings.
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
