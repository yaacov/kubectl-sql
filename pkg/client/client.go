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
	"fmt"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// Client provides information required to query the kubernetes server.
type Client struct {
	Config *rest.Config
}

// List resources by resource name.
func (c Client) List(resourceName string) ([]unstructured.Unstructured, error) {
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
	list, err := res.List(v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, err
}

// Look for a resource matching request resource name.
func (c Client) getResourceGroupVersion(resourceName string) (v1.APIResource, string, string, error) {
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
		return v1.APIResource{}, "", "", fmt.Errorf("Failed to find resource")
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
