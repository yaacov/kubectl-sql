package eval

import (
	"github.com/yaacov/tree-search-language/v6/pkg/walkers/semantics"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// EvalFunctionFactory build an evaluation method for one item that returns a value using a key.
func EvalFunctionFactory(item unstructured.Unstructured) semantics.EvalFunc {
	return func(key string) (interface{}, bool) {
		return ExtractValue(item, key)
	}
}
