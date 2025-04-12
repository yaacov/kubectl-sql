package eval

import (
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestExtractValue(t *testing.T) {
	creationTime := time.Now().UTC().Truncate(time.Second)
	deletionTime := creationTime.Add(time.Hour)

	item := unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":              "test-pod",
				"namespace":         "default",
				"creationTimestamp": creationTime.Format(time.RFC3339),
				"deletionTimestamp": deletionTime.Format(time.RFC3339),
				"labels": map[string]interface{}{
					"app": "test",
				},
				"annotations": map[string]interface{}{
					"note": "test annotation",
				},
			},
			"spec": map[string]interface{}{
				"replicas": int64(3),
				"nested": map[string]interface{}{
					"value": "nested-value",
				},
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "container1",
						"image": "nginx:latest",
						"ports": []interface{}{
							map[string]interface{}{
								"containerPort": int64(80),
								"protocol":      "TCP",
							},
							map[string]interface{}{
								"containerPort": int64(443),
								"protocol":      "TCP",
							},
						},
						"resources": map[string]interface{}{
							"limits": map[string]interface{}{
								"cpu":    "500m",
								"memory": "512Mi",
							},
						},
					},
					map[string]interface{}{
						"name":  "container2",
						"image": "redis:latest",
						"ports": []interface{}{
							map[string]interface{}{
								"containerPort": int64(6379),
								"protocol":      "TCP",
							},
						},
					},
				},
				"volumes": []interface{}{
					map[string]interface{}{
						"name": "data",
						"configMap": map[string]interface{}{
							"name": "config-data",
						},
					},
				},
			},
			"status": map[string]interface{}{
				"phase": "Running",
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
					},
					map[string]interface{}{
						"type":   "PodScheduled",
						"status": "True",
					},
				},
				"podIP":     "10.0.0.1",
				"hostIP":    "192.168.1.1",
				"ready":     true,
				"startTime": creationTime.Add(time.Minute).Format(time.RFC3339),
				"containerStatuses": []interface{}{
					map[string]interface{}{
						"name":         "container1",
						"ready":        true,
						"restartCount": int64(0),
						"started":      true,
					},
					map[string]interface{}{
						"name":         "container2",
						"ready":        true,
						"restartCount": int64(2),
						"started":      true,
					},
				},
				"metrics": map[string]interface{}{
					"cpu":    map[string]interface{}{"usage": "250m"},
					"memory": map[string]interface{}{"usage": "256Mi"},
				},
				"numericValues": []interface{}{1, 2, 3, 4, 5},
				"mixedArray": []interface{}{
					"string",
					42,
					true,
					map[string]interface{}{"key": "value"},
					[]interface{}{1, 2, 3},
				},
			},
		},
	}

	tests := []struct {
		name     string
		key      string
		want     interface{}
		wantBool bool
	}{
		{"name", "name", "test-pod", true},
		{"namespace", "namespace", "default", true},
		{"created", "created", creationTime.UTC(), true},
		{"deleted", "deleted", deletionTime.UTC(), true},
		{"label", "labels.app", "test", true},
		{"annotation", "annotations.note", "test annotation", true},
		{"nested spec", "spec.nested.value", "nested-value", true},
		{"replicas", "spec.replicas", float64(3), true},
		{"non-existent", "invalid.path", nil, true},

		// Test array indexing
		{"container name", "spec.containers[0].name", "container1", true},
		{"container image", "spec.containers[0].image", "nginx:latest", true},
		{"second container", "spec.containers[1].name", "container2", true},

		// Test nested arrays
		{"container port", "spec.containers[0].ports[0].containerPort", float64(80), true},
		{"second port", "spec.containers[0].ports[1].containerPort", float64(443), true},

		// Test complex nested objects
		{"resource limits", "spec.containers[0].resources.limits.cpu", "500m", true},
		{"volume configmap", "spec.volumes[0].configMap.name", "config-data", true},

		// Test booleans
		{"pod ready", "status.ready", true, true},
		{"container ready", "status.containerStatuses[0].ready", true, true},

		// Test status fields
		{"pod phase", "status.phase", "Running", true},
		{"pod IP", "status.podIP", "10.0.0.1", true},

		// Test conditions
		{"condition type", "status.conditions[0].type", "Ready", true},

		// Test more complex jsonpath expressions
		{"all container names", "spec.containers[*].name", []interface{}{"container1", "container2"}, true},
		{"all container ports", "spec.containers[0].ports[*].containerPort", []interface{}{float64(80), float64(443)}, true},
		{"restart counts", "status.containerStatuses[*].restartCount", []interface{}{float64(0), float64(2)}, true},

		// Test numeric arrays
		{"numeric values", "status.numericValues", []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5)}, true},
		{"first numeric value", "status.numericValues[0]", float64(1), true},

		// Test mixed arrays
		{"mixed array string", "status.mixedArray[0]", "string", true},
		{"mixed array number", "status.mixedArray[1]", float64(42), true},
		{"mixed array boolean", "status.mixedArray[2]", true, true},
		{"mixed array object", "status.mixedArray[3].key", "value", true},

		// Test deep nesting
		{"metrics cpu", "status.metrics.cpu.usage", "250m", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ExtractValue(item, tt.key)

			// Use reflect.DeepEqual for comparing values, especially arrays/slices
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractValue() got = %v (type %T), want %v (type %T)", got, got, tt.want, tt.want)
			}
			if got1 != tt.wantBool {
				t.Errorf("extractValue() got1 = %v, want %v", got1, tt.wantBool)
			}
		})
	}
}
