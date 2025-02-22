package eval

import (
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ExtractValue(item, tt.key)
			if got != tt.want {
				t.Errorf("extractValue() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.wantBool {
				t.Errorf("extractValue() got1 = %v, want %v", got1, tt.wantBool)
			}
		})
	}
}
