package query

import (
	"reflect"
	"testing"
)

func TestSortItems(t *testing.T) {
	tests := []struct {
		name      string
		items     []map[string]interface{}
		queryOpts *QueryOptions
		want      []map[string]interface{}
	}{
		{
			name: "simple ascending",
			items: []map[string]interface{}{
				{"a": 2}, {"a": 1}, {"a": 3},
			},
			queryOpts: &QueryOptions{
				OrderBy:    []OrderOption{{Field: SelectOption{Field: ".a", Alias: "a"}, Descending: false}},
				HasOrderBy: true,
			},
			want: []map[string]interface{}{
				{"a": 1}, {"a": 2}, {"a": 3},
			},
		},
		{
			name: "with alias descending",
			items: []map[string]interface{}{
				{"x": 5}, {"x": 3}, {"x": 4},
			},
			queryOpts: &QueryOptions{
				Select:     []SelectOption{{Field: ".x", Alias: "y"}},
				OrderBy:    []OrderOption{{Field: SelectOption{Field: ".x", Alias: "y"}, Descending: true}},
				HasSelect:  true,
				HasOrderBy: true,
			},
			want: []map[string]interface{}{
				{"x": 5}, {"x": 4}, {"x": 3},
			},
		},
		{
			name: "with sum reducer ascending",
			items: []map[string]interface{}{
				{"nums": []interface{}{1, 3}},
				{"nums": []interface{}{2, 3}},
			},
			queryOpts: &QueryOptions{
				Select:     []SelectOption{{Field: ".nums", Alias: "nums", Reducer: "sum"}},
				OrderBy:    []OrderOption{{Field: SelectOption{Field: ".nums", Alias: "nums", Reducer: "sum"}, Descending: false}},
				HasSelect:  true,
				HasOrderBy: true,
			},
			want: []map[string]interface{}{
				{"nums": []interface{}{1, 3}},
				{"nums": []interface{}{2, 3}},
			},
		},
		{
			name: "with len reducer ascending",
			items: []map[string]interface{}{
				{"arr": []interface{}{1, 2}},
				{"arr": []interface{}{1}},
				{"arr": []interface{}{1, 2, 3}},
			},
			queryOpts: &QueryOptions{
				Select:     []SelectOption{{Field: ".arr", Alias: "arr", Reducer: "len"}},
				OrderBy:    []OrderOption{{Field: SelectOption{Field: ".arr", Alias: "arr", Reducer: "len"}, Descending: false}},
				HasSelect:  true,
				HasOrderBy: true,
			},
			want: []map[string]interface{}{
				{"arr": []interface{}{1}},
				{"arr": []interface{}{1, 2}},
				{"arr": []interface{}{1, 2, 3}},
			},
		},
		{
			name: "with any reducer descending",
			items: []map[string]interface{}{
				{"flags": []interface{}{false, false, true}},
				{"flags": []interface{}{false, false}},
				{"flags": []interface{}{true, false}},
			},
			queryOpts: &QueryOptions{
				Select:     []SelectOption{{Field: ".flags", Alias: "flags", Reducer: "any"}},
				OrderBy:    []OrderOption{{Field: SelectOption{Field: ".flags", Alias: "flags", Reducer: "any"}, Descending: true}},
				HasSelect:  true,
				HasOrderBy: true,
			},
			want: []map[string]interface{}{
				{"flags": []interface{}{false, false, true}},
				{"flags": []interface{}{true, false}},
				{"flags": []interface{}{false, false}},
			},
		},
		{
			name: "with all reducer ascending",
			items: []map[string]interface{}{
				{"flags": []interface{}{true, true}},
				{"flags": []interface{}{true, false}},
				{"flags": []interface{}{}},
			},
			queryOpts: &QueryOptions{
				Select:     []SelectOption{{Field: ".flags", Alias: "flags", Reducer: "all"}},
				OrderBy:    []OrderOption{{Field: SelectOption{Field: ".flags", Alias: "flags", Reducer: "all"}, Descending: false}},
				HasSelect:  true,
				HasOrderBy: true,
			},
			want: []map[string]interface{}{
				{"flags": []interface{}{true, false}},
				{"flags": []interface{}{}},
				{"flags": []interface{}{true, true}},
			},
		},
		{
			name: "compound sort",
			items: []map[string]interface{}{
				{"a": 1, "b": 1},
				{"a": 1, "b": 2},
				{"a": 0, "b": 3},
			},
			queryOpts: &QueryOptions{
				OrderBy: []OrderOption{
					{Field: SelectOption{Field: ".a", Alias: "a"}, Descending: false},
					{Field: SelectOption{Field: ".b", Alias: "b"}, Descending: true},
				},
				HasOrderBy: true,
			},
			want: []map[string]interface{}{
				{"a": 0, "b": 3},
				{"a": 1, "b": 2},
				{"a": 1, "b": 1},
			},
		},
	}

	for _, tc := range tests {
		got, err := SortItems(tc.items, tc.queryOpts)
		if err != nil {
			t.Errorf("%q: unexpected error: %v", tc.name, err)
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%q: got %v, want %v", tc.name, got, tc.want)
		}
	}
}
