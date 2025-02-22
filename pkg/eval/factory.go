package eval

import (
	"strings"

	"github.com/yaacov/tree-search-language/v6/pkg/walkers/semantics"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// EvalFunctionFactory build an evaluation method for one item that returns a value using a key.
func EvalFunctionFactory(item unstructured.Unstructured) semantics.EvalFunc {
	return func(key string) (interface{}, bool) {
		return ExtractValue(item, key)
	}
}

// JoinEvalFunctionFactory build an evaluation method for two items that returns a value using a key.
func JoinEvalFunctionFactory(item1, item2 unstructured.Unstructured, prefix1, prefix2 string) semantics.EvalFunc {
	return func(key string) (interface{}, bool) {
		// Use item1 if has prefix1
		if strings.HasPrefix(key, prefix1+".") {
			return ExtractValue(item1, strings.TrimPrefix(key, prefix1+"."))
		}

		// Use item2 if has prefix2
		if strings.HasPrefix(key, prefix2+".") {
			return ExtractValue(item2, strings.TrimPrefix(key, prefix2+"."))
		}

		// Default to use item2
		return ExtractValue(item2, key)
	}
}
