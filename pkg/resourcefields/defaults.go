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
*/

package resourcefields

import (
	"github.com/yaacov/kubectl-sql/pkg/query"
)

// GetDefaultSelectFields returns the default fields to display when SELECT * is used
func GetDefaultSelectFields(resourceName string) []query.SelectOption {
	// Define resource-specific fields based on resource type
	switch resourceName {
	case "pods", "pod", "po":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "metadata.namespace", Alias: "Namespace"},
			{Field: "status.phase", Alias: "Status"},
			{Field: "spec.nodeName", Alias: "Node"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	case "services", "service", "svc":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "metadata.namespace", Alias: "Namespace"},
			{Field: "spec.type", Alias: "Type"},
			{Field: "spec.clusterIP", Alias: "ClusterIP"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	case "deployments", "deployment", "deploy":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "metadata.namespace", Alias: "Namespace"},
			{Field: "status.replicas", Alias: "Replicas"},
			{Field: "status.availableReplicas", Alias: "Available"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	case "nodes", "node", "no":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "status.conditions[?(@.type==\"Ready\")].status", Alias: "Ready"},
			{Field: "status.nodeInfo.kubeletVersion", Alias: "Version"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	case "configmaps", "configmap", "cm":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "metadata.namespace", Alias: "Namespace"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	case "secrets", "secret":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "metadata.namespace", Alias: "Namespace"},
			{Field: "type", Alias: "Type"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	case "persistentvolumeclaims", "persistentvolumeclaim", "pvc":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "metadata.namespace", Alias: "Namespace"},
			{Field: "status.phase", Alias: "Status"},
			{Field: "spec.volumeName", Alias: "Volume"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	// KubeVirt resources
	case "virtualmachines", "virtualmachine", "vm":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "metadata.namespace", Alias: "Namespace"},
			{Field: "spec.running", Alias: "Running"},
			{Field: "status.ready", Alias: "Ready"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	case "virtualmachineinstances", "virtualmachineinstance", "vmi":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "metadata.namespace", Alias: "Namespace"},
			{Field: "status.phase", Alias: "Phase"},
			{Field: "status.nodeName", Alias: "Node"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	case "virtualmachineinstancereplicasets", "virtualmachineinstancereplicaset", "vmirs":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "metadata.namespace", Alias: "Namespace"},
			{Field: "spec.replicas", Alias: "Replicas"},
			{Field: "status.readyReplicas", Alias: "Ready"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	case "virtualmachineinstancemigrations", "virtualmachineinstancemigration", "vmim":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "metadata.namespace", Alias: "Namespace"},
			{Field: "spec.vmName", Alias: "VM"},
			{Field: "status.phase", Alias: "Phase"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	// CDI resources
	case "datavolumes", "datavolume", "dv":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "metadata.namespace", Alias: "Namespace"},
			{Field: "status.phase", Alias: "Phase"},
			{Field: "spec.source.http.url", Alias: "Source"},
			{Field: "spec.pvc.storageClassName", Alias: "StorageClass"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	case "cdis", "cdi":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "status.phase", Alias: "Phase"},
			{Field: "status.conditions[?(@.type==\"Available\")].status", Alias: "Available"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	case "cdiconfigs", "cdiconfig":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "spec.scratchSpaceStorageClass", Alias: "ScratchClass"},
			{Field: "spec.uploadProxyURLOverride", Alias: "UploadURL"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	// Forklift resources
	case "providers", "provider":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "metadata.namespace", Alias: "Namespace"},
			{Field: "spec.type", Alias: "Type"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	case "plans", "plan":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "metadata.namespace", Alias: "Namespace"},
			{Field: "spec.provider.source.name", Alias: "Source"},
			{Field: "spec.provider.destination.name", Alias: "Destination"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	case "migrations", "migration":
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "metadata.namespace", Alias: "Namespace"},
			{Field: "spec.plan.name", Alias: "Plan"},
			{Field: "status.phase", Alias: "Phase"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
		}
	default:
		// Default fields for unknown resource types
		return []query.SelectOption{
			{Field: "metadata.name", Alias: "Name"},
			{Field: "metadata.namespace", Alias: "Namespace"},
			{Field: "metadata.creationTimestamp", Alias: "Created"},
			{Field: "status.phase", Alias: "Phase"},
		}
	}
}
