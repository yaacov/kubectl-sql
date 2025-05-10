package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/yaacov/kubectl-sql/pkg/query"
)

// NamePrinter prints only the names of resources
type NamePrinter struct {
	items        []map[string]interface{}
	writer       io.Writer
	nameField    string
	queryOptions *query.QueryOptions
}

// NewNamePrinter creates a new NamePrinter
func NewNamePrinter() *NamePrinter {
	return &NamePrinter{
		items:     []map[string]interface{}{},
		writer:    os.Stdout,
		nameField: "metadata.name",
	}
}

// WithWriter sets the output writer
func (n *NamePrinter) WithWriter(writer io.Writer) *NamePrinter {
	n.writer = writer
	return n
}

// WithNameField sets the JSON path to the name field
func (n *NamePrinter) WithNameField(path string) *NamePrinter {
	n.nameField = path
	return n
}

// WithQueryOptions sets the query options for name extraction
func (n *NamePrinter) WithQueryOptions(queryOptions *query.QueryOptions) *NamePrinter {
	n.queryOptions = queryOptions
	return n
}

// AddItem adds an item to the output
func (n *NamePrinter) AddItem(item map[string]interface{}) *NamePrinter {
	n.items = append(n.items, item)
	return n
}

// AddItems adds multiple items to the output
func (n *NamePrinter) AddItems(items []map[string]interface{}) *NamePrinter {
	n.items = append(n.items, items...)
	return n
}

// extractName extracts the name from an item
func (n *NamePrinter) extractName(item map[string]interface{}) string {
	if n.nameField == "" {
		return ""
	}

	// Use query.GetValue if queryOptions are set, otherwise fallback to GetValueByPathString
	// This allows for
	//   - alias names in the queryOptions
	//   - reducer methods, len, sum, any, and all
	if n.queryOptions != nil {
		// Debug output before GetValue call
		if n.queryOptions.DebugLevel > 0 {
			debugInfo := map[string]interface{}{
				"operation":       "GetValue",
				"source":          "name.go:extractName",
				"nameField":       n.nameField,
				"hasQueryOptions": true,
			}
			debugJSON, _ := json.MarshalIndent(debugInfo, "", "  ")
			fmt.Fprintf(os.Stderr, "DEBUG: Before GetValue call:\n%s\n", string(debugJSON))
		}

		val, err := query.GetValue(item, n.nameField, n.queryOptions)
		if err != nil {
			return ""
		}
		return valueToString(val)
	}

	value, err := query.GetValueByPathString(item, n.nameField, n.queryOptions.DebugLevel)
	if err != nil {
		return ""
	}

	return valueToString(value)
}

// Print outputs just the names of all items
func (n *NamePrinter) Print() error {
	for _, item := range n.items {
		name := n.extractName(item)
		if name != "" {
			fmt.Fprintln(n.writer, name)
		}
	}
	return nil
}

// PrintEmpty outputs a message when there are no items
func (n *NamePrinter) PrintEmpty(message string) error {
	if message != "" {
		fmt.Fprintln(n.writer, message)
	}
	return nil
}
