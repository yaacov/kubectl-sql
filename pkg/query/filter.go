package query

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yaacov/tree-search-language/v6/pkg/tsl"
	"github.com/yaacov/tree-search-language/v6/pkg/walkers/semantics"
)

// ParseWhereClause parses a WHERE clause string into a TSL tree
func ParseWhereClause(whereClause string) (*tsl.TSLNode, error) {
	tree, err := tsl.ParseTSL(whereClause)
	if err != nil {
		return nil, fmt.Errorf("failed to parse where clause: %v", err)
	}

	return tree, nil
}

// ApplyFilter filters items using a TSL tree
func ApplyFilter(items []map[string]interface{}, tree *tsl.TSLNode, queryOpts *QueryOptions) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// Filter the items collection using the TSL tree
	for _, item := range items {
		eval := evalFactory(item, queryOpts)

		matchingFilter, err := semantics.Walk(tree, eval)
		if err != nil {
			// Ignore errors for items that don't match the filter
			// and set the matchingFilter to false
			matchingFilter = false
		}

		// Convert interface{} to bool
		if match, ok := matchingFilter.(bool); ok && match {
			results = append(results, item)
		}
	}

	return results, nil
}

// evalFactory gets an item and returns a method that will get the field and return its value
func evalFactory(item map[string]interface{}, queryOpts *QueryOptions) semantics.EvalFunc {
	return func(k string) (interface{}, bool) {
		// Debug output before GetValue call
		if queryOpts.DebugLevel > 0 {
			debugInfo := map[string]interface{}{
				"operation": "GetValue",
				"source":    "filter.go:evalFactory",
				"key":       k,
				"filter":    true,
				"context":   "WHERE clause evaluation",
			}
			debugJSON, _ := json.MarshalIndent(debugInfo, "", "  ")
			fmt.Fprintf(os.Stderr, "DEBUG: Before GetValue call in filter:\n%s\n", string(debugJSON))
		}

		// Use GetValue to respect aliases and reducers
		if v, err := GetValue(item, k, queryOpts); err == nil {
			return v, true
		}
		return nil, true
	}
}
