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
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/yaml.v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/urfave/cli/v2"
)

func actionsGet(c *cli.Context) error {
	namespace := ""

	// Get user request.
	resourceName, query := getRequst(c)

	// Get kubeconfig.
	kubeconfig := getKubeConfig(c)
	config, err := kubeconfig.ClientConfig()
	errExit("Failed to load client conifg", err)

	// Get resource information.
	resource, resourceList := getResource(config, resourceName)
	group, version := getGroupVersion(resourceList)
	if resource.Namespaced {
		namespace = getNamespace(c, kubeconfig)
	}

	// Get dynamic client.
	client, err := dynamic.NewForConfig(config)
	errExit("Failed to create rest client", err)

	// Get all resource objects.
	res := client.Resource(schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource.Name,
	})
	list, err := res.List(metav1.ListOptions{})
	errExit("Failed to list objects", err)

	// Print selected objects.
	printItems(c, list, namespace, query)

	return nil
}

// Get user quary.
func getRequst(c *cli.Context) (string, string) {
	resourceName := c.String("resource")
	query := c.String("query")

	// Parse command args
	if len(resourceName) == 0 && c.Args().Len() == 1 {
		resourceName = c.Args().Get(0)
	} else if len(query) == 0 && c.Args().Len() == 3 && c.Args().Get(1) == "where" {
		resourceName = c.Args().Get(0)
		query = c.Args().Get(2)
	}

	if len(resourceName) == 0 {
		errExit("Failed to parse resource query", fmt.Errorf("missing resource name or query"))
	}

	return resourceName, query
}

func printItemYAML(item unstructured.Unstructured) {
	yaml, err := yaml.Marshal(item)
	errExit("Failed to marshal item", err)

	fmt.Printf("\n%+v\n", string(yaml))
}

func printItemJSON(item unstructured.Unstructured) {
	yaml, err := json.Marshal(item)
	errExit("Failed to marshal item", err)

	fmt.Printf("\n%+v\n", string(yaml))
}

func printItemTableRaw(item unstructured.Unstructured) {
	fmt.Printf("%-30s %-30s %-20s\n", item.GetNamespace(), item.GetName(), item.GetCreationTimestamp().Format(time.RFC3339))
}
