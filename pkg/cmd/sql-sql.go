package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yaacov/kubectl-sql/pkg/printers"
)

// isValidFieldIdentifier checks if a field name matches the allowed pattern
func isValidFieldIdentifier(field string) bool {
	// Matches patterns like:
	// - simple: name, first_name, my.field
	// - array access: items[0], my.array[123]
	pattern := `^[a-zA-Z_]([a-zA-Z0-9_.]*(?:\[\d+\])?)*$`
	match, _ := regexp.MatchString(pattern, field)
	return match
}

// CompleteSQL parses SQL query into components
func (o *SQLOptions) CompleteSQL(query string) error {
	// Convert to uppercase for case-insensitive matching
	upperQuery := strings.ToUpper(query)

	if !strings.HasPrefix(upperQuery, "SELECT") {
		return fmt.Errorf("query must start with SELECT")
	}

	// Extract SELECT fields
	fromIndex := strings.Index(upperQuery, "FROM")
	if fromIndex == -1 {
		return fmt.Errorf("missing FROM clause in query")
	}
	selectFields := strings.TrimSpace(query[6:fromIndex])

	// Find WHERE and ON clauses if they exist
	whereIndex := strings.Index(upperQuery, "WHERE")
	if fromIndex == -1 {
		return fmt.Errorf("missing WHERE clause in query")
	}

	onIndex := strings.Index(upperQuery, "ON")

	// Extract resources from FROM clause
	fromPart := query[fromIndex+4 : whereIndex]
	if onIndex != -1 {
		fromPart = query[fromIndex+4 : onIndex]
	}

	resources := strings.Split(strings.TrimSpace(fromPart), ",")
	for i, r := range resources {
		resources[i] = strings.TrimSpace(r)
	}

	// Validate number of resources based on presence of ON clause
	if onIndex != -1 {
		if len(resources) != 2 {
			return fmt.Errorf("when using ON clause, exactly two resources must be specified")
		}
	} else {
		if len(resources) != 1 {
			return fmt.Errorf("without ON clause, exactly one resource must be specified")
		}
	}

	o.requestedResources = resources

	// Extract ON clause if it exists
	if onIndex != -1 {
		onPart := query[onIndex+2 : whereIndex]
		o.requestedOnQuery = strings.TrimSpace(onPart)

		// Validate ON clause
		if o.requestedOnQuery == "" {
			return fmt.Errorf("ON clause cannot be empty")
		}
	}

	// Extract WHERE clause if it exists
	if whereIndex != -1 {
		wherePart := query[whereIndex+5:]
		o.requestedQuery = strings.TrimSpace(wherePart)

		// Validate WHERE clause
		if o.requestedQuery == "" && o.requestedOnQuery == "" {
			return fmt.Errorf("WHERE clause cannot be empty")
		}
	}

	// Read SQL plugin specific configurations.
	if err := o.readConfigFile(o.requestedSQLConfigPath); err != nil {
		return err
	}

	// Process SELECT fields if not "*"
	if selectFields != "*" {
		if len(strings.TrimSpace(selectFields)) == 0 {
			return fmt.Errorf("SELECT clause cannot be empty")
		}

		// Split fields and create new table fields map
		fields := strings.Split(selectFields, ",")
		newTableFields := make(printers.TableFieldsMap)

		// Validate each field
		for _, field := range fields {
			field = strings.TrimSpace(field)
			if !isValidFieldIdentifier(field) {
				return fmt.Errorf("invalid field identifier: %s", field)
			}
		}

		// For each resource, create table fields with the selected columns
		for _, resource := range o.requestedResources {
			tableFields := make([]printers.TableField, 0, len(fields))
			for _, field := range fields {
				field = strings.TrimSpace(field)
				tableFields = append(tableFields, printers.TableField{
					Title: field,
					Name:  field,
				})
			}
			newTableFields[resource] = tableFields
		}
		o.defaultTableFields = newTableFields
	}

	return nil
}
