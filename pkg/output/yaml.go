package output

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

// YAMLPrinter prints data as YAML
type YAMLPrinter struct {
	items  []map[string]interface{}
	writer io.Writer
}

// NewYAMLPrinter creates a new YAMLPrinter
func NewYAMLPrinter() *YAMLPrinter {
	return &YAMLPrinter{
		items:  []map[string]interface{}{},
		writer: os.Stdout,
	}
}

// WithWriter sets the output writer
func (y *YAMLPrinter) WithWriter(writer io.Writer) *YAMLPrinter {
	y.writer = writer
	return y
}

// AddItem adds an item to the YAML output
func (y *YAMLPrinter) AddItem(item map[string]interface{}) *YAMLPrinter {
	y.items = append(y.items, item)
	return y
}

// AddItems adds multiple items to the YAML output
func (y *YAMLPrinter) AddItems(items []map[string]interface{}) *YAMLPrinter {
	y.items = append(y.items, items...)
	return y
}

// Print outputs the items as YAML
func (y *YAMLPrinter) Print() error {
	data, err := yaml.Marshal(y.items)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %v", err)
	}

	_, err = fmt.Fprintln(y.writer, string(data))
	return err
}

// PrintEmpty outputs an empty YAML document or a message when there are no items
func (y *YAMLPrinter) PrintEmpty(message string) error {
	var data []byte
	var err error

	if message == "" {
		// If no message, just print an empty array
		data, err = yaml.Marshal([]interface{}{})
	} else {
		// If message provided, print only the message
		result := map[string]interface{}{
			"message": message,
		}
		data, err = yaml.Marshal(result)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %v", err)
	}

	_, err = fmt.Fprintln(y.writer, string(data))
	return err
}
