package query

import (
	"reflect"
	"testing"
	"time"
)

func TestApplyFilter(t *testing.T) {
	tests := []struct {
		name       string
		items      []map[string]interface{}
		where      string
		selectOpts []SelectOption
		want       []map[string]interface{}
		wantErr    bool
	}{
		{
			name: "numeric greater than",
			items: []map[string]interface{}{
				{"foo": 1}, {"foo": 2}, {"foo": 3},
			},
			where:      "foo>1",
			selectOpts: nil,
			want: []map[string]interface{}{
				{"foo": 2}, {"foo": 3},
			},
		},
		{
			name: "string equality",
			items: []map[string]interface{}{
				{"bar": "a"}, {"bar": "b"}, {"bar": "a"},
			},
			where:      `bar = "a"`,
			selectOpts: nil,
			want: []map[string]interface{}{
				{"bar": "a"}, {"bar": "a"},
			},
		},
		{
			name: "boolean truthy",
			items: []map[string]interface{}{
				{"baz": true}, {"baz": false}, {"baz": true},
			},
			where:      "baz",
			selectOpts: nil,
			want: []map[string]interface{}{
				{"baz": true}, {"baz": true},
			},
		},
		{
			name: "logical AND",
			items: []map[string]interface{}{
				{"foo": 2, "bar": 3}, {"foo": 3, "bar": 5}, {"foo": 1, "bar": 2},
			},
			where:      "foo > 1 and bar < 4",
			selectOpts: nil,
			want: []map[string]interface{}{
				{"foo": 2, "bar": 3},
			},
		},
		{
			name: "logical OR",
			items: []map[string]interface{}{
				{"foo": 1}, {"foo": 3}, {"foo": 5},
			},
			where:      "foo < 2 or foo > 4",
			selectOpts: nil,
			want: []map[string]interface{}{
				{"foo": 1}, {"foo": 5},
			},
		},
		{
			name: "nested field equality",
			items: []map[string]interface{}{
				{"n": map[string]interface{}{"x": 1.0}}, {"n": map[string]interface{}{"x": 2.0}},
			},
			where:      "n.x = 1.0",
			selectOpts: nil,
			want: []map[string]interface{}{
				{"n": map[string]interface{}{"x": 1.0}},
			},
		},
		{
			name: "parentheses grouping",
			items: []map[string]interface{}{
				{"a": 1, "b": 2, "c": false},
				{"a": 2, "b": 3, "c": false},
				{"a": 3, "b": 1, "c": true},
			},
			where:      "(a > 1 and b < 2) or c",
			selectOpts: nil,
			want: []map[string]interface{}{
				{"a": 3, "b": 1, "c": true},
			},
		},
		{
			name: "alias usage",
			items: []map[string]interface{}{
				{"foo": 1}, {"foo": 3},
			},
			where:      "f>2",
			selectOpts: []SelectOption{{Field: ".foo", Alias: "f"}},
			want: []map[string]interface{}{
				{"foo": 3},
			},
		},
		{
			name: "reducer alias",
			items: []map[string]interface{}{
				{"nums": []interface{}{1, 2}},
				{"nums": []interface{}{3, 4}},
			},
			where:      "s>5",
			selectOpts: []SelectOption{{Field: ".nums", Alias: "s", Reducer: "sum"}},
			want: []map[string]interface{}{
				{"nums": []interface{}{3, 4}},
			},
		},
		{
			name: "int-float equality",
			items: []map[string]interface{}{
				{"foo": 3}, {"foo": 3.0}, {"foo": 4},
			},
			where:      "foo = 3.0",
			selectOpts: nil,
			want: []map[string]interface{}{
				{"foo": 3}, {"foo": 3.0},
			},
		},
		{
			name: "date-string equality",
			items: []map[string]interface{}{
				{"date": time.Date(2020, time.January, 2, 15, 4, 5, 0, time.UTC)},
				{"date": time.Date(2021, time.January, 2, 15, 4, 5, 0, time.UTC)},
			},
			where:      `date = "2020-01-02T15:04:05Z"`,
			selectOpts: nil,
			want: []map[string]interface{}{
				{"date": time.Date(2020, time.January, 2, 15, 4, 5, 0, time.UTC)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := ParseWhereClause(tt.where)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseWhereClause(%q) error = %v, wantErr %v", tt.where, err, tt.wantErr)
			}
			if err != nil {
				return
			}

			// Create a QueryOptions struct with the select options
			queryOpts := &QueryOptions{
				Select: tt.selectOpts,
				Where:  tt.where,
			}

			got, err := ApplyFilter(tt.items, tree, queryOpts)
			if err != nil {
				t.Fatalf("ApplyFilter error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}
