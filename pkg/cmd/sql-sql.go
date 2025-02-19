package cmd

import (
	"fmt"
	"strings"
)

// CompleteSQL parses SQL query into components
func (o *SQLOptions) CompleteSQL(query string) error {
	// Convert to uppercase for case-insensitive matching
	upperQuery := strings.ToUpper(query)

	// Extract FROM clause
	fromIndex := strings.Index(upperQuery, "FROM")
	if fromIndex == -1 {
		return fmt.Errorf("missing FROM clause in query")
	}

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

	return nil
}
