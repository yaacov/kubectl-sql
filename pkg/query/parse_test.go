package query

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseQueryString(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected *QueryOptions
		err      bool
		errMsg   string
	}{
		{
			name:  "empty query",
			query: "",
			expected: &QueryOptions{
				Select:     nil,
				HasSelect:  false,
				Where:      "",
				OrderBy:    nil,
				HasOrderBy: false,
				Limit:      -1,
				HasLimit:   false,
				From:       FromOptions{},
				HasFrom:    false,
			},
		},
		{
			name:  "simple select and alias",
			query: "SELECT foo, bar as baz",
			expected: &QueryOptions{
				Select: []SelectOption{
					{Field: ".foo", Alias: "", Reducer: ""},
					{Field: ".bar", Alias: "baz", Reducer: ""},
				},
				HasSelect:  true,
				Where:      "",
				OrderBy:    nil,
				HasOrderBy: false,
				Limit:      -1,
				HasLimit:   false,
				From:       FromOptions{},
				HasFrom:    false,
			},
		},
		{
			name:  "where and limit",
			query: "where count>0 limit 5",
			expected: &QueryOptions{
				Select:     nil,
				HasSelect:  false,
				Where:      "count>0",
				OrderBy:    nil,
				HasOrderBy: false,
				Limit:      5,
				HasLimit:   true,
				From:       FromOptions{},
				HasFrom:    false,
			},
		},
		{
			name:  "order by asc and desc",
			query: "order by foo desc, bar ASC",
			expected: &QueryOptions{
				Select:    nil,
				HasSelect: false,
				Where:     "",
				OrderBy: []OrderOption{
					{
						Field:      SelectOption{Field: ".foo", Alias: "foo", Reducer: ""},
						Descending: true,
					},
					{
						Field:      SelectOption{Field: ".bar", Alias: "bar", Reducer: ""},
						Descending: false,
					},
				},
				HasOrderBy: true,
				Limit:      -1,
				HasLimit:   false,
				From:       FromOptions{},
				HasFrom:    false,
			},
		},
		{
			name:  "combined full query",
			query: "SELECT sum(x) as total, y WHERE y>1 ORDER BY x DESC, y LIMIT 10",
			expected: &QueryOptions{
				Select: []SelectOption{
					{Field: ".x", Alias: "total", Reducer: "sum"},
					{Field: ".y", Alias: "", Reducer: ""},
				},
				HasSelect: true,
				Where:     "y>1",
				OrderBy: []OrderOption{
					{
						Field:      SelectOption{Field: ".x", Alias: "total", Reducer: "sum"},
						Descending: true,
					},
					{
						Field:      SelectOption{Field: ".y", Alias: "", Reducer: ""},
						Descending: false,
					},
				},
				HasOrderBy: true,
				Limit:      10,
				HasLimit:   true,
				From:       FromOptions{},
				HasFrom:    false,
			},
		},
		{
			name:  "order by alias",
			query: "SELECT foo as f, bar as b ORDER BY f DESC, b",
			expected: &QueryOptions{
				Select: []SelectOption{
					{Field: ".foo", Alias: "f", Reducer: ""},
					{Field: ".bar", Alias: "b", Reducer: ""},
				},
				HasSelect: true,
				Where:     "",
				OrderBy: []OrderOption{
					{
						Field:      SelectOption{Field: ".foo", Alias: "f", Reducer: ""},
						Descending: true,
					},
					{
						Field:      SelectOption{Field: ".bar", Alias: "b", Reducer: ""},
						Descending: false,
					},
				},
				HasOrderBy: true,
				Limit:      -1,
				HasLimit:   false,
				From:       FromOptions{},
				HasFrom:    false,
			},
		},
		{
			name:  "simple from",
			query: "SELECT foo FROM pods",
			expected: &QueryOptions{
				Select: []SelectOption{
					{Field: ".foo", Alias: "", Reducer: ""},
				},
				HasSelect:  true,
				Where:      "",
				OrderBy:    nil,
				HasOrderBy: false,
				Limit:      -1,
				HasLimit:   false,
				From: FromOptions{
					Name: "pods",
				},
				HasFrom: true,
			},
		},
		{
			name:  "from with namespace/name syntax",
			query: "SELECT foo FROM default/pods WHERE count>0",
			expected: &QueryOptions{
				Select: []SelectOption{
					{Field: ".foo", Alias: "", Reducer: ""},
				},
				HasSelect:  true,
				Where:      "count>0",
				OrderBy:    nil,
				HasOrderBy: false,
				Limit:      -1,
				HasLimit:   false,
				From: FromOptions{
					Name:      "pods",
					Namespace: "default",
				},
				HasFrom: true,
			},
		},
		{
			name:  "from with */name syntax (all namespaces)",
			query: "SELECT foo FROM */pods WHERE count>0",
			expected: &QueryOptions{
				Select: []SelectOption{
					{Field: ".foo", Alias: "", Reducer: ""},
				},
				HasSelect:  true,
				Where:      "count>0",
				OrderBy:    nil,
				HasOrderBy: false,
				Limit:      -1,
				HasLimit:   false,
				From: FromOptions{
					Name:          "pods",
					AllNamespaces: true,
				},
				HasFrom: true,
			},
		},
		{
			name:  "complete query with namespace/name format",
			query: "SELECT foo, bar FROM kube-system/deployments WHERE foo>10 ORDER BY bar DESC LIMIT 5",
			expected: &QueryOptions{
				Select: []SelectOption{
					{Field: ".foo", Alias: "", Reducer: ""},
					{Field: ".bar", Alias: "", Reducer: ""},
				},
				HasSelect: true,
				Where:     "foo>10",
				OrderBy: []OrderOption{
					{
						Field:      SelectOption{Field: ".bar", Alias: "", Reducer: ""},
						Descending: true,
					},
				},
				HasOrderBy: true,
				Limit:      5,
				HasLimit:   true,
				From: FromOptions{
					Name:      "deployments",
					Namespace: "kube-system",
				},
				HasFrom: true,
			},
		},
		{
			name:  "valid select fields with complex paths",
			query: "SELECT metadata.name, status.phase, spec.containers[0].image",
			expected: &QueryOptions{
				Select: []SelectOption{
					{Field: ".metadata.name", Alias: "", Reducer: ""},
					{Field: ".status.phase", Alias: "", Reducer: ""},
					{Field: ".spec.containers[0].image", Alias: "", Reducer: ""},
				},
				HasSelect:  true,
				Where:      "",
				OrderBy:    nil,
				HasOrderBy: false,
				Limit:      -1,
				HasLimit:   false,
				From:       FromOptions{},
				HasFrom:    false,
			},
		},
		{
			name:  "valid select fields with all features",
			query: "SELECT metadata.name as name, count(status.conditions[*]) as condition_count, spec.replicas",
			expected: &QueryOptions{
				Select: []SelectOption{
					{Field: ".metadata.name", Alias: "name", Reducer: ""},
					{Field: ".status.conditions[*]", Alias: "condition_count", Reducer: "count"},
					{Field: ".spec.replicas", Alias: "", Reducer: ""},
				},
				HasSelect:  true,
				Where:      "",
				OrderBy:    nil,
				HasOrderBy: false,
				Limit:      -1,
				HasLimit:   false,
				From:       FromOptions{},
				HasFrom:    false,
			},
		},
		{
			name:  "valid select with wildcards",
			query: "SELECT metadata.labels[*], status.containerStatuses[*].ready",
			expected: &QueryOptions{
				Select: []SelectOption{
					{Field: ".metadata.labels[*]", Alias: "", Reducer: ""},
					{Field: ".status.containerStatuses[*].ready", Alias: "", Reducer: ""},
				},
				HasSelect:  true,
				Where:      "",
				OrderBy:    nil,
				HasOrderBy: false,
				Limit:      -1,
				HasLimit:   false,
				From:       FromOptions{},
				HasFrom:    false,
			},
		},
		{
			name:   "invalid reducer name",
			query:  "SELECT invalid_function(name) FROM pods",
			err:    true,
			errMsg: "unsupported reducer function: invalid_function",
		},
		{
			name:   "invalid select syntax - missing field after function",
			query:  "SELECT count() FROM pods",
			err:    true,
			errMsg: "invalid select clause format",
		},
		{
			name:   "invalid select syntax - empty field name",
			query:  "SELECT FROM pods",
			err:    true,
			errMsg: "invalid select clause",
		},
		{
			name:   "invalid select syntax - unbalanced parentheses",
			query:  "SELECT count(items FROM pods",
			err:    true,
			errMsg: "unbalanced parentheses in select clause",
		},
		// Error cases for FROM validation
		{
			name:   "from with empty resource name",
			query:  "SELECT foo FROM ",
			err:    true,
			errMsg: "FROM clause requires a resource name",
		},
		{
			name:   "from with invalid resource name",
			query:  "SELECT foo FROM invalid_name",
			err:    true,
			errMsg: "resource name 'invalid_name' is invalid",
		},
		{
			name:   "from with invalid namespace",
			query:  "SELECT foo FROM invalid_namespace/pods",
			err:    true,
			errMsg: "namespace name 'invalid_namespace' is invalid",
		},
		{
			name:   "from with too long resource name",
			query:  "SELECT foo FROM " + strings.Repeat("a", 64),
			err:    true,
			errMsg: "resource name '" + strings.Repeat("a", 64) + "' is too long",
		},
		{
			name:   "from with too long namespace",
			query:  "SELECT foo FROM " + strings.Repeat("a", 64) + "/pods",
			err:    true,
			errMsg: "namespace name '" + strings.Repeat("a", 64) + "' is too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQueryString(tt.query)

			// Check error status
			if (err != nil) != tt.err {
				t.Fatalf("unexpected error status: %v", err)
			}

			// Check error message if expected
			if tt.err && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error message doesn't contain expected text.\nExpected substring: %s\nActual error: %s", tt.errMsg, err.Error())
				}
			}

			// Skip deep comparison if we expected an error
			if !tt.err && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ParseQueryString(%q) =\n  %#v\nexpected\n  %#v", tt.query, got, tt.expected)
			}
		})
	}
}

// Additional tests for the name validation function
func TestValidateK8sName(t *testing.T) {
	validNames := []string{
		"my-name",
		"app123",
		"kubernetes",
		"a",
		"my.name.with.dots",
		"123-456",
	}

	invalidNames := []string{
		"",
		"My-Name",               // uppercase not allowed
		"-leading-dash",         // can't start with dash
		"trailing-dash-",        // can't end with dash
		"name_with_underscore",  // underscores not allowed
		"name with spaces",      // spaces not allowed
		strings.Repeat("a", 64), // too long
	}

	for _, name := range validNames {
		if err := validateK8sName(name, "test"); err != nil {
			t.Errorf("validateK8sName(%q) returned error for valid name: %v", name, err)
		}
	}

	for _, name := range invalidNames {
		if err := validateK8sName(name, "test"); err == nil {
			t.Errorf("validateK8sName(%q) did not return error for invalid name", name)
		}
	}
}

// Add test for the validateFromOptions function
func TestValidateFromOptions(t *testing.T) {
	tests := []struct {
		name    string
		from    FromOptions
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid options - name only",
			from: FromOptions{
				Name: "pods",
			},
			wantErr: false,
		},
		{
			name: "valid options - with namespace",
			from: FromOptions{
				Name:      "pods",
				Namespace: "default",
			},
			wantErr: false,
		},
		{
			name: "valid options - with asterisk namespace",
			from: FromOptions{
				Name:      "pods",
				Namespace: "*",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			from: FromOptions{
				Name:      "",
				Namespace: "default",
			},
			wantErr: true,
			errMsg:  "resource name cannot be empty",
		},
		{
			name: "invalid name",
			from: FromOptions{
				Name:      "Invalid_Name",
				Namespace: "default",
			},
			wantErr: true,
			errMsg:  "resource name 'Invalid_Name' is invalid",
		},
		{
			name: "invalid namespace",
			from: FromOptions{
				Name:      "pods",
				Namespace: "Invalid_Namespace",
			},
			wantErr: true,
			errMsg:  "namespace name 'Invalid_Namespace' is invalid",
		},
		{
			name: "empty namespace is valid",
			from: FromOptions{
				Name:      "pods",
				Namespace: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFromOptions(tt.from)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFromOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error message doesn't contain expected text.\nExpected substring: %s\nActual error: %s",
					tt.errMsg, err.Error())
			}
		})
	}
}

func TestParseSelectClauseFunctionOptionalParentheses(t *testing.T) {
	tests := []struct {
		input string
		want  SelectOption
	}{
		{"len hello", SelectOption{Field: ".hello", Reducer: "len", Alias: ""}},
		{"len(hello)", SelectOption{Field: ".hello", Reducer: "len", Alias: ""}},
		{"sum value as total", SelectOption{Field: ".value", Reducer: "sum", Alias: "total"}},
		{"sum(value) as total", SelectOption{Field: ".value", Reducer: "sum", Alias: "total"}},
	}

	for _, tc := range tests {
		got := parseSelectClause(tc.input)
		if len(got) != 1 {
			t.Errorf("parseSelectClause(%q) returned %d opts, want 1", tc.input, len(got))
			continue
		}
		if !reflect.DeepEqual(got[0], tc.want) {
			t.Errorf("parseSelectClause(%q)[0] = %+v, want %+v", tc.input, got[0], tc.want)
		}
	}
}

// TestParseQueryStringIrregular tests parsing of more complex, irregular query patterns
// that involve array paths, nested structures, and pattern matching.
func TestParseQueryStringIrregular(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected *QueryOptions
		err      bool
		errMsg   string
	}{
		{
			name:  "array path with pattern matching",
			query: "SELECT name, (spec.template.spec.containers[*].image) FROM */deployments WHERE any (spec.template.spec.containers[*].image ~= 'quay.*')",
			expected: &QueryOptions{
				Select: []SelectOption{
					{Field: ".name", Alias: "", Reducer: ""},
					{Field: ".spec.template.spec.containers[*].image", Alias: "", Reducer: ""},
				},
				HasSelect:  true,
				Where:      "any (spec.template.spec.containers[*].image ~= 'quay.*')",
				OrderBy:    nil,
				HasOrderBy: false,
				Limit:      -1,
				HasLimit:   false,
				From: FromOptions{
					Name:          "deployments",
					AllNamespaces: true,
				},
				HasFrom: true,
			},
		},
		{
			name:  "len function with array path",
			query: "SELECT name, len(spec.template.spec.containers[*].image) FROM */deployments WHERE any (spec.template.spec.containers[*].image ~= 'quay.*')",
			expected: &QueryOptions{
				Select: []SelectOption{
					{Field: ".name", Alias: "", Reducer: ""},
					{Field: ".spec.template.spec.containers[*].image", Alias: "", Reducer: "len"},
				},
				HasSelect:  true,
				Where:      "any (spec.template.spec.containers[*].image ~= 'quay.*')",
				OrderBy:    nil,
				HasOrderBy: false,
				Limit:      -1,
				HasLimit:   false,
				From: FromOptions{
					Name:          "deployments",
					AllNamespaces: true,
				},
				HasFrom: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseQueryString(tt.query)

			// Check error status
			if (err != nil) != tt.err {
				t.Fatalf("unexpected error status: %v", err)
			}

			// Check error message if expected
			if tt.err && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error message doesn't contain expected text.\nExpected substring: %s\nActual error: %s", tt.errMsg, err.Error())
				}
			}

			// Skip deep comparison if we expected an error
			if !tt.err {
				if !reflect.DeepEqual(got, tt.expected) {
					t.Errorf("ParseQueryString(%q) produced different result than expected:", tt.query)

					// Compare HasSelect flag
					if got.HasSelect != tt.expected.HasSelect {
						t.Errorf("  HasSelect: got %v, expected %v", got.HasSelect, tt.expected.HasSelect)
					}

					// Compare Select fields
					if !reflect.DeepEqual(got.Select, tt.expected.Select) {
						t.Errorf("  Select fields mismatch:")
						t.Errorf("    got: %#v", got.Select)
						t.Errorf("    expected: %#v", tt.expected.Select)
					}

					// Compare Where clause
					if got.Where != tt.expected.Where {
						t.Errorf("  Where: got %q, expected %q", got.Where, tt.expected.Where)
					}

					// Compare HasOrderBy flag
					if got.HasOrderBy != tt.expected.HasOrderBy {
						t.Errorf("  HasOrderBy: got %v, expected %v", got.HasOrderBy, tt.expected.HasOrderBy)
					}

					// Compare OrderBy fields
					if !reflect.DeepEqual(got.OrderBy, tt.expected.OrderBy) {
						t.Errorf("  OrderBy fields mismatch:")
						t.Errorf("    got: %#v", got.OrderBy)
						t.Errorf("    expected: %#v", tt.expected.OrderBy)
					}

					// Compare HasLimit flag and Limit value
					if got.HasLimit != tt.expected.HasLimit {
						t.Errorf("  HasLimit: got %v, expected %v", got.HasLimit, tt.expected.HasLimit)
					}
					if got.Limit != tt.expected.Limit {
						t.Errorf("  Limit: got %d, expected %d", got.Limit, tt.expected.Limit)
					}

					// Compare HasFrom flag
					if got.HasFrom != tt.expected.HasFrom {
						t.Errorf("  HasFrom: got %v, expected %v", got.HasFrom, tt.expected.HasFrom)
					}

					// Compare From options
					if !reflect.DeepEqual(got.From, tt.expected.From) {
						t.Errorf("  From options mismatch:")
						t.Errorf("    got: %#v", got.From)
						t.Errorf("    expected: %#v", tt.expected.From)
					}
				}
			}
		})
	}
}
