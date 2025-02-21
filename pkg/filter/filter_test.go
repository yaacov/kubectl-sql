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
