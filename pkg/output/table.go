package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/yaacov/kubectl-sql/pkg/query"
)

// Header represents a table column header with display text and a JSON path
type Header struct {
	DisplayName  string
	JSONPath     string
	SelectOption *query.SelectOption
}

// TablePrinter prints tabular data with dynamically sized columns
type TablePrinter struct {
	headers      []Header
	items        []map[string]interface{}
	padding      int
	minWidth     int
	writer       io.Writer
	maxColWidth  int
	expandedData map[int]string // Stores expanded data for each row by index
	noHeaders    bool           // Controls whether to display headers
	debugLevel   int            // Debug level for logging
}

// NewTablePrinter creates a new TablePrinter
func NewTablePrinter() *TablePrinter {
	return &TablePrinter{
		headers:      []Header{},
		items:        []map[string]interface{}{},
		padding:      2,
		minWidth:     10,
		writer:       os.Stdout,
		maxColWidth:  50, // Prevent very wide columns
		expandedData: make(map[int]string),
		noHeaders:    false, // Default to showing headers
		debugLevel:   0,     // Default debug level
	}
}

// WithHeaders sets the table headers with display names and JSON paths
func (t *TablePrinter) WithHeaders(headers ...Header) *TablePrinter {
	t.headers = headers
	return t
}

// WithoutHeaders configures the printer to not display headers
func (t *TablePrinter) WithoutHeaders() *TablePrinter {
	t.noHeaders = true
	return t
}

// WithPadding sets the padding between columns
func (t *TablePrinter) WithPadding(padding int) *TablePrinter {
	t.padding = padding
	return t
}

// WithMinWidth sets the minimum column width
func (t *TablePrinter) WithMinWidth(minWidth int) *TablePrinter {
	t.minWidth = minWidth
	return t
}

// WithMaxWidth sets the maximum column width
func (t *TablePrinter) WithMaxWidth(maxWidth int) *TablePrinter {
	t.maxColWidth = maxWidth
	return t
}

// WithWriter sets the output writer
func (t *TablePrinter) WithWriter(writer io.Writer) *TablePrinter {
	t.writer = writer
	return t
}

// WithExpandedData sets expanded data for a specific row index
func (t *TablePrinter) WithExpandedData(index int, data string) *TablePrinter {
	t.expandedData[index] = data
	return t
}

// WithDebugLevel sets the debug level for the table printer
func (t *TablePrinter) WithDebugLevel(level int) *TablePrinter {
	t.debugLevel = level
	return t
}

// AddItem adds an item to the table
func (t *TablePrinter) AddItem(item map[string]interface{}) *TablePrinter {
	t.items = append(t.items, item)
	return t
}

// AddItemWithExpanded adds an item to the table with expanded data
func (t *TablePrinter) AddItemWithExpanded(item map[string]interface{}, expanded string) *TablePrinter {
	index := len(t.items)
	t.items = append(t.items, item)
	t.expandedData[index] = expanded
	return t
}

// AddItems adds multiple items to the table
func (t *TablePrinter) AddItems(items []map[string]interface{}) *TablePrinter {
	t.items = append(t.items, items...)
	return t
}

// extractValue extracts a value from an item using a JSON path
func (t *TablePrinter) extractValue(item map[string]interface{}, header Header) string {
	if header.JSONPath == "" {
		// No path provided, return empty string
		return ""
	}

	// Use SelectOption if available, otherwise use JSONPath directly
	if header.SelectOption != nil {
		// Create a special QueryOptions with just this SelectOption
		specialQueryOptions := &query.QueryOptions{
			Select:     []query.SelectOption{*header.SelectOption},
			HasSelect:  true,
			DebugLevel: t.debugLevel,
		}

		// Debug output before GetValue call if debug is enabled
		if t.debugLevel > 0 {
			debugInfo := map[string]interface{}{
				"operation":    "GetValue",
				"source":       "table.go:extractValue",
				"selectOption": *header.SelectOption,
				"field":        header.SelectOption.Field,
				"alias":        header.SelectOption.Alias,
				"reducer":      header.SelectOption.Reducer,
				"debugLevel":   t.debugLevel,
			}
			debugJSON, _ := json.MarshalIndent(debugInfo, "", "  ")
			fmt.Fprintf(os.Stderr, "DEBUG: Before GetValue call with SelectOption:\n%s\n", string(debugJSON))
		}

		val, err := query.GetValue(item, header.JSONPath, specialQueryOptions)
		if err != nil {
			return ""
		}
		return valueToString(val)
	}

	value, err := query.GetValueByPathString(item, header.JSONPath, t.debugLevel)
	if err != nil {
		return ""
	}

	return valueToString(value)
}

// valueToString converts a value of any supported type to a string
func valueToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case float32:
		return fmt.Sprintf("%g", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case []interface{}, []string, []int, []float64:
		// Handle arrays by converting to JSON
		jsonBytes, err := json.Marshal(v)
		if err == nil {
			return string(jsonBytes)
		}
		// Fallback to default formatting if JSON marshaling fails
		return fmt.Sprintf("%v", v)
	default:
		// For other types, use default string conversion
		return fmt.Sprintf("%v", v)
	}
}

// calculateColumnWidths determines the optimal width for each column
func (t *TablePrinter) calculateColumnWidths() []int {
	numCols := len(t.headers)
	if numCols == 0 {
		return []int{}
	}

	// Initialize widths with minimum values
	widths := make([]int, numCols)
	for i := range widths {
		widths[i] = t.minWidth
	}

	// Check header widths
	for i, header := range t.headers {
		headerWidth := utf8.RuneCountInString(header.DisplayName)
		if headerWidth > widths[i] {
			widths[i] = min(headerWidth, t.maxColWidth)
		}
	}

	// Calculate row data for width determination
	for _, item := range t.items {
		for i, header := range t.headers {
			value := t.extractValue(item, header)
			cellWidth := utf8.RuneCountInString(value)
			if cellWidth > widths[i] {
				widths[i] = min(cellWidth, t.maxColWidth)
			}
		}
	}

	return widths
}

// Print prints the table with dynamic column widths
func (t *TablePrinter) Print() error {
	widths := t.calculateColumnWidths()
	if len(widths) == 0 {
		return nil
	}

	// Print headers
	if !t.noHeaders {
		headerRow := make([]string, len(t.headers))
		for i, header := range t.headers {
			headerRow[i] = header.DisplayName
		}
		t.printRow(headerRow, widths)
	}

	// Print item rows and expanded data if available
	for i, item := range t.items {
		row := make([]string, len(t.headers))
		for j, header := range t.headers {
			row[j] = t.extractValue(item, header)
		}
		t.printRow(row, widths)

		// Print expanded data if it exists for this row
		if expanded, exists := t.expandedData[i]; exists && expanded != "" {
			// Split expanded data into lines and add prefix
			lines := strings.Split(expanded, "\n")
			for _, line := range lines {
				fmt.Fprintf(t.writer, "  │ %s\n", line)
			}
		}
	}

	return nil
}

// PrintEmpty prints a message when there are no items to display
func (t *TablePrinter) PrintEmpty(message string) error {
	fmt.Fprintln(t.writer, message)
	return nil
}

// printRow prints a single row with the specified column widths
func (t *TablePrinter) printRow(row []string, widths []int) {
	var sb strings.Builder

	// Don't print empty rows
	if len(row) == 0 {
		return
	}

	// Print each cell with padding
	for i, cell := range row {
		if i >= len(widths) {
			break
		}

		// Truncate if the cell is too long
		displayCell := cell
		if utf8.RuneCountInString(cell) > t.maxColWidth {
			displayCell = cell[:t.maxColWidth-3] + "..."
		}

		// Format with proper padding
		format := fmt.Sprintf("%%-%ds", widths[i]+t.padding)
		sb.WriteString(fmt.Sprintf(format, displayCell))
	}

	fmt.Fprintln(t.writer, strings.TrimRight(sb.String(), " "))
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
