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

package eval

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// extractValue extract a value from an item using a key.
func extractValue(item unstructured.Unstructured, key string) (interface{}, bool) {
	// Check for reserved words.
	switch key {
	case "name":
		return item.GetName(), true
	case "namespace":
		return item.GetNamespace(), true
	case "created":
		return item.GetCreationTimestamp().Time, true
	case "deleted":
		return item.GetDeletionTimestamp().Time, true
	}

	// Check for labels and annotations.
	if len(key) > 7 && key[:7] == "labels." {
		value, ok := item.GetLabels()[key[7:]]
		return handleMetadataValue(value, ok)
	}

	if len(key) > 12 && key[:12] == "annotations." {
		value, ok := item.GetAnnotations()[key[12:]]
		return handleMetadataValue(value, ok)
	}

	// Check for numbers, booleans, dates and strings.
	object, ok := getNestedObject(item.Object, key)
	if !ok {
		return nil, true
	}

	return convertObjectToValue(object)
}

func handleMetadataValue(value string, exists bool) (interface{}, bool) {
	if !exists {
		return nil, true
	}
	if len(value) == 0 {
		return true, true
	}
	return stringValue(value), true
}

func convertObjectToValue(object interface{}) (interface{}, bool) {
	switch v := object.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case string:
		return stringValue(v), true
	}
	return nil, true
}
