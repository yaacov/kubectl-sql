package eval

import (
	"reflect"
	"testing"
)

func TestGetNestedObject(t *testing.T) {
	testObj := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "value",
			},
		},
		"items": []interface{}{
			map[string]interface{}{"name": "item1"},
			map[string]interface{}{"name": "item2"},
		},
	}

	tests := []struct {
		name     string
		obj      interface{}
		key      string
		want     interface{}
		wantBool bool
	}{
		{
			name:     "simple nested access",
			obj:      testObj,
			key:      "a.b.c",
			want:     "value",
			wantBool: true,
		},
		{
			name:     "array access by index",
			obj:      testObj,
			key:      "items[1]",
			want:     map[string]interface{}{"name": "item1"},
			wantBool: true,
		},
		{
			name:     "array access by index (2)",
			obj:      testObj,
			key:      "items[2]",
			want:     map[string]interface{}{"name": "item2"},
			wantBool: true,
		},
		{
			name:     "array access with dot notation",
			obj:      testObj,
			key:      "items.1",
			want:     map[string]interface{}{"name": "item1"},
			wantBool: true,
		},
		{
			name:     "invalid path",
			obj:      testObj,
			key:      "x.y.z",
			want:     nil,
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotBool := getNestedObject(tt.obj, tt.key)
			if !reflect.DeepEqual(got, tt.want) || gotBool != tt.wantBool {
				t.Errorf("getNestedObject() = (%v, %v), want (%v, %v)", got, gotBool, tt.want, tt.wantBool)
			}
		})
	}
}

func TestSplitKeys(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want []string
	}{
		{
			name: "simple dot notation",
			key:  "a.b.c",
			want: []string{"a", "b", "c"},
		},
		{
			name: "array notation",
			key:  "items[1].name",
			want: []string{"items", "1", "name"},
		},
		{
			name: "mixed notation",
			key:  "spec.containers[1].name",
			want: []string{"spec", "containers", "1", "name"},
		},
		{
			name: "array access with dot notation",
			key:  "items.1.name",
			want: []string{"items", "1", "name"},
		},
		{
			name: "multiple array accesses with dot notation",
			key:  "pods.2.containers.0.name",
			want: []string{"pods", "2", "containers", "0", "name"},
		},
		{
			name: "mixed array access notation",
			key:  "pods.2.containers[0].name",
			want: []string{"pods", "2", "containers", "0", "name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitKeys(tt.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
