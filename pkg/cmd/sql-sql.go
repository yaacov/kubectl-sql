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

// isValidK8sResourceName checks if a resource name follows Kubernetes naming conventions
func isValidK8sResourceName(resource string) bool {
	// Matches lowercase words separated by dots or slashes
	// Examples: pods, deployments, apps/v1/deployments
	pattern := `^[a-z]+([a-z0-9-]*[a-z0-9])?(/[a-z0-9]+)*$`
	match, _ := regexp.MatchString(pattern, resource)
	return match
}

// QueryType represents the type of SQL query
type QueryType int

const (
	SimpleQuery QueryType = iota
	JoinQuery
	JoinWhereQuery
)

// parseFields extracts and validates SELECT fields
func (o *SQLOptions) parseFields(selectFields string) error {
	if selectFields == "*" {
		return nil
	}

	if len(strings.TrimSpace(selectFields)) == 0 {
		return fmt.Errorf("SELECT clause cannot be empty")
	}

	fields := strings.Split(selectFields, ",")
	tableFields := make([]printers.TableField, 0, len(fields))

	for _, field := range fields {
		field = strings.TrimSpace(field)

		// Check for AS syntax
		parts := strings.Split(strings.ToUpper(field), " AS ")
		var name, title string

		if len(parts) == 2 {
			// We have an AS clause
			name = strings.TrimSpace(field[:strings.Index(strings.ToUpper(field), " AS ")])
			title = strings.TrimSpace(field[strings.Index(strings.ToUpper(field), " AS ")+4:])

			if !isValidFieldIdentifier(name) {
				return fmt.Errorf("invalid field identifier before AS: %s", name)
			}
			if !isValidFieldIdentifier(title) {
				return fmt.Errorf("invalid field identifier after AS: %s", title)
			}
		} else {
			// No AS clause, use field as both name and title
			if !isValidFieldIdentifier(field) {
				return fmt.Errorf("invalid field identifier: %s", field)
			}
			name = field
			title = field
		}

		// Append to table fields
		tableFields = append(tableFields, printers.TableField{
			Name:  name,
			Title: title,
		})

		// Append to default aliases
		o.defaultAliases[title] = name
	}

	o.defaultTableFields[printers.SelectedFields] = tableFields
	return nil
}

// parseResources validates and sets the requested resources
func (o *SQLOptions) parseResources(resources []string, queryType QueryType) error {
	for i, r := range resources {
		r = strings.TrimSpace(r)

		// Split resource on "/" to check for namespace
		parts := strings.Split(r, "/")
		var resourceName string

		switch len(parts) {
		case 1:
			resourceName = parts[0]
		case 2:
			o.namespace = parts[0]
			resourceName = parts[1]
		default:
			return fmt.Errorf("invalid resource format: %s, expected [namespace/]resource", r)
		}

		if !isValidK8sResourceName(resourceName) {
			return fmt.Errorf("invalid resource name: %s", resourceName)
		}

		resources[i] = resourceName
	}

	if queryType == SimpleQuery && len(resources) != 1 {
		return fmt.Errorf("without ON clause, exactly one resource must be specified")
	}
	if (queryType == JoinQuery || queryType == JoinWhereQuery) && len(resources) != 2 {
		return fmt.Errorf("when using ON clause, exactly two resources must be specified")
	}

	o.requestedResources = resources
	return nil
}

// identifyQueryType determines the type of SQL query and returns relevant indices
func (o *SQLOptions) identifyQueryType(query string) (QueryType, map[string]int, error) {
	upperQuery := strings.ToUpper(query)
	if !strings.HasPrefix(upperQuery, "SELECT") {
		return SimpleQuery, nil, fmt.Errorf("query must start with SELECT")
	}

	indices := map[string]int{
		"SELECT": 0,
		"FROM":   strings.Index(upperQuery, " FROM "),
		"JOIN":   strings.Index(upperQuery, " JOIN "),
		"ON":     strings.Index(upperQuery, " ON "),
		"WHERE":  strings.Index(upperQuery, " WHERE "),
	}

	if indices["FROM"] == -1 {
		return 0, nil, fmt.Errorf("missing FROM clause in query")
	}

	if indices["JOIN"] == -1 {
		return SimpleQuery, indices, nil
	}

	if indices["ON"] == -1 {
		return 0, nil, fmt.Errorf("JOIN clause requires ON condition")
	}

	if indices["WHERE"] == -1 {
		return JoinQuery, indices, nil
	}

	return JoinWhereQuery, indices, nil
}

// parseQueryParts extracts and validates different parts of the query
func (o *SQLOptions) parseQueryParts(query string, indices map[string]int, queryType QueryType) error {
	// Parse FROM resource (only one resource allowed)
	var fromEnd int
	if indices["JOIN"] != -1 {
		fromEnd = indices["JOIN"]
	} else if indices["WHERE"] != -1 {
		fromEnd = indices["WHERE"]
	} else {
		fromEnd = len(query)
	}

	fromPart := strings.TrimSpace(query[indices["FROM"]+5 : fromEnd])
	resources := strings.Split(fromPart, ",")
	if len(resources) != 1 {
		return fmt.Errorf("only one resource allowed in FROM clause")
	}

	// If JOIN exists, add the joined resource
	var allResources []string
	if queryType != SimpleQuery {
		joinStart := indices["JOIN"] + 5
		joinEnd := indices["ON"]
		joinResource := strings.TrimSpace(query[joinStart:joinEnd])
		allResources = []string{resources[0], joinResource}
	} else {
		allResources = []string{resources[0]}
	}

	if err := o.parseResources(allResources, queryType); err != nil {
		return err
	}

	// Parse SELECT fields
	selectFields := strings.TrimSpace(query[6:indices["FROM"]])
	if err := o.parseFields(selectFields); err != nil {
		return err
	}

	// Parse ON clause for JOIN queries
	if queryType != SimpleQuery {
		onStart := indices["ON"] + 3
		onEnd := indices["WHERE"]
		if onEnd == -1 {
			onEnd = len(query)
		}
		o.requestedOnQuery = strings.TrimSpace(query[onStart:onEnd])
		if o.requestedOnQuery == "" {
			return fmt.Errorf("ON clause cannot be empty")
		}
	}

	// Parse WHERE clause if present
	if indices["WHERE"] != -1 {
		wherePart := strings.TrimSpace(query[indices["WHERE"]+6:])
		if wherePart == "" {
			return fmt.Errorf("WHERE clause cannot be empty")
		}
		o.requestedQuery = wherePart
	}

	return nil
}

// CompleteSQL parses SQL query into components
func (o *SQLOptions) CompleteSQL(query string) error {
	// Read SQL plugin specific configurations
	err := o.readConfigFile(o.requestedSQLConfigPath)
	if err != nil {
		return err
	}

	queryType, indices, err := o.identifyQueryType(query)
	if err != nil {
		return err
	}

	if err := o.parseQueryParts(query, indices, queryType); err != nil {
		return err
	}

	return nil
}
