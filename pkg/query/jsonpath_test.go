package query

import (
	"reflect"
	"testing"
	"time"
)

func TestGetValueByPathString(t *testing.T) {
	obj := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "pod1",
			"labels": map[string]interface{}{
				"env": "prod",
			},
		},
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{"name": "c1"},
				map[string]interface{}{"name": "c2"},
			},
		},
		"intVal":   42,
		"floatVal": 3.14,
		"boolVal":  true,
		"date":     time.Date(2020, time.January, 2, 15, 4, 5, 0, time.UTC),
	}

	tests := []struct {
		name    string
		path    string
		want    interface{}
		wantErr bool
	}{
		{"simple dot", "metadata.name", "pod1", false},
		{"leading dot", ".metadata.name", "pod1", false},
		{"braces", "{{ metadata.name }}", "pod1", false},
		{"map key", "metadata.labels[env]", "prod", false},
		{"array index", "spec.containers[1].name", "c2", false},
		{"wildcard", "spec.containers[*].name", []interface{}{"c1", "c2"}, false},
		{"bool value", "boolVal", true, false},
		{"int value", "intVal", 42, false},
		{"float value", "floatVal", 3.14, false},
		{"date value", "date", time.Date(2020, time.January, 2, 15, 4, 5, 0, time.UTC), false},
		{"missing field", "nonexistent.field", nil, false},
		{"out of bounds", "spec.containers[10].name", nil, true},
		{"bad index", "spec.containers[foo].name", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetValueByPathString(obj, tt.path, 0)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("path %q = %#v, want %#v", tt.path, got, tt.want)
			}
		})
	}
}

func TestGetValue(t *testing.T) {
	obj := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "pod1",
			"labels": map[string]interface{}{
				"env": "prod",
			},
		},
		"nums":  []interface{}{1, 2, 3},
		"flags": []interface{}{true, false},
	}

	tests := []struct {
		name    string
		field   string
		opts    []SelectOption
		want    interface{}
		wantErr bool
	}{
		{"direct", "metadata.name", nil, "pod1", false},
		{"alias lookup", "aliasEnv", []SelectOption{
			{Alias: "aliasEnv", Field: "metadata.labels[env]"},
		}, "prod", false},
		{"unknown name", "foo", nil, nil, false},
		{"sum reducer", "sumNums", []SelectOption{
			{Alias: "sumNums", Field: ".nums", Reducer: "sum"},
		}, float64(6), false},
		{"len reducer", "lenNums", []SelectOption{
			{Alias: "lenNums", Field: ".nums", Reducer: "len"},
		}, 3, false},
		{"any reducer", "anyFlags", []SelectOption{
			{Alias: "anyFlags", Field: ".flags", Reducer: "any"},
		}, true, false},
		{"all reducer", "allFlags", []SelectOption{
			{Alias: "allFlags", Field: ".flags", Reducer: "all"},
		}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a QueryOptions struct with the SelectOption slice
			queryOpts := &QueryOptions{
				Select: tt.opts,
			}

			got, err := GetValue(obj, tt.field, queryOpts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetValue err = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetValue(%q) = %#v, want %#v", tt.field, got, tt.want)
			}
		})
	}
}
