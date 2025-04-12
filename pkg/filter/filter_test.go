package filter

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestFilter(t *testing.T) {
	items := []unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "test1",
					"labels": map[string]interface{}{
						"app": "web",
					},
				},
				"spec": map[string]interface{}{
					"replicas": int64(3),
					"containers": []interface{}{
						map[string]interface{}{
							"name": "nginx",
							"ports": []interface{}{
								map[string]interface{}{
									"containerPort": int64(80),
								},
								map[string]interface{}{
									"containerPort": int64(443),
								},
							},
						},
						map[string]interface{}{
							"name": "sidecar",
							"ports": []interface{}{
								map[string]interface{}{
									"containerPort": int64(8080),
								},
							},
						},
					},
				},
			},
		},
		{
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": "test2",
					"labels": map[string]interface{}{
						"app": "db",
					},
				},
				"spec": map[string]interface{}{
					"replicas": int64(1),
					"containers": []interface{}{
						map[string]interface{}{
							"name": "postgres",
							"ports": []interface{}{
								map[string]interface{}{
									"containerPort": int64(5432),
								},
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name      string
		query     string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "filter by name",
			query:     "name = 'test1'",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "filter by label",
			query:     "labels.app = 'web'",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "filter by replicas",
			query:     "spec.replicas > 2",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "invalid query",
			query:     "invalid query",
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "filter with any on array element",
			query:     "any (spec.containers[*].name = 'nginx')",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "filter with any on nested array",
			query:     "any (spec.containers[*].ports[*].containerPort < 400)",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "filter with all on array element",
			query:     "all (spec.containers[*].ports[*].containerPort < 9000)",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "filter with array count",
			query:     "len (spec.containers) > 1",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "filter comparing array values",
			query:     "'postgres' in spec.containers[*].name",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "invalid query",
			query:     "invalid query",
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Query: tt.query,
				CheckColumnName: func(s string) (string, error) {
					return s, nil
				},
			}

			got, err := c.Filter(items)
			if (err != nil) != tt.wantErr {
				t.Errorf("Filter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantCount {
				t.Errorf("Filter() got = %v items, want %v", len(got), tt.wantCount)
			}
		})
	}
}
