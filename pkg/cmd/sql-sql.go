package cmd

import (
	"fmt"
	"regexp"
	"strconv"
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

// isValidNamespace checks if a namespace name is valid according to Kubernetes naming conventions
// or if it's the special "*" value for all namespaces
func isValidNamespace(namespace string) bool {
	// Special case for "all namespaces"
	if namespace == "*" {
		return true
	}

	pattern := `^[a-z0-9]([a-z0-9\-]*[a-z0-9])?$`
	match, _ := regexp.MatchString(pattern, namespace)
	return match
}

// QueryType represents the type of SQL query
type QueryType int

const (
	SimpleQuery QueryType = iota
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
			// Check for namespace validity
			namespace := parts[0]
			if !isValidNamespace(namespace) {
				return fmt.Errorf("invalid namespace: %s", namespace)
			}

			// Set namespace options
			if namespace == "*" {
				o.allNamespaces = true
			} else {
				o.namespace = namespace
			}
			resourceName = parts[1]
		default:
			return fmt.Errorf("invalid resource format: %s, expected [namespace/]resource or */resource for all namespaces", r)
		}

		if !isValidK8sResourceName(resourceName) {
			return fmt.Errorf("invalid resource name: %s", resourceName)
		}

		resources[i] = resourceName
	}

	if len(resources) != 1 {
		return fmt.Errorf("exactly one resource must be specified")
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
		"SELECT":   0,
		"FROM":     strings.Index(upperQuery, " FROM "),
		"JOIN":     strings.Index(upperQuery, " JOIN "),
		"ON":       strings.Index(upperQuery, " ON "),
		"WHERE":    strings.Index(upperQuery, " WHERE "),
		"ORDER BY": strings.Index(upperQuery, " ORDER BY "),
		"LIMIT":    strings.Index(upperQuery, " LIMIT "),
	}

	if indices["FROM"] == -1 {
		return 0, nil, fmt.Errorf("missing FROM clause in query")
	}

	return SimpleQuery, indices, nil
}

// parseOrderBy extracts and validates the ORDER BY clause
func (o *SQLOptions) parseOrderBy(query string, indices map[string]int) error {
	if indices["ORDER BY"] == -1 {
		return nil
	}

	orderByStart := indices["ORDER BY"] + 9
	var orderByEnd int
	if indices["LIMIT"] != -1 {
		orderByEnd = indices["LIMIT"]
	} else {
		orderByEnd = len(query)
	}

	orderByStr := strings.TrimSpace(query[orderByStart:orderByEnd])
	if orderByStr == "" {
		return fmt.Errorf("ORDER BY clause cannot be empty")
	}

	fields := strings.Split(orderByStr, ",")
	orderByFields := make([]printers.OrderByField, 0, len(fields))

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		parts := strings.Fields(field)
		if len(parts) == 0 {
			continue
		}

		fieldName := parts[0]
		// Check for possible alias
		if alias, err := o.checkColumnName(fieldName); err == nil {
			fieldName = alias
		}

		orderBy := printers.OrderByField{
			Name:       fieldName,
			Descending: false,
		}

		// Check for DESC/ASC modifier
		if len(parts) > 1 && strings.ToUpper(parts[1]) == "DESC" {
			orderBy.Descending = true
		}

		orderByFields = append(orderByFields, orderBy)
	}

	o.orderByFields = orderByFields
	return nil
}

// parseLimit extracts and validates the LIMIT clause
func (o *SQLOptions) parseLimit(query string, indices map[string]int) error {
	if indices["LIMIT"] == -1 {
		return nil
	}

	limitStart := indices["LIMIT"] + 6
	limitStr := strings.TrimSpace(query[limitStart:])

	// Check if there are other clauses after LIMIT
	if space := strings.Index(limitStr, " "); space != -1 {
		limitStr = limitStr[:space]
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return fmt.Errorf("invalid LIMIT value: %s", limitStr)
	}

	if limit < 0 {
		return fmt.Errorf("LIMIT cannot be negative: %d", limit)
	}

	o.limit = limit
	return nil
}

// parseQueryParts extracts and validates different parts of the query
func (o *SQLOptions) parseQueryParts(query string, indices map[string]int, queryType QueryType) error {
	// Parse FROM resource (only one resource allowed)
	var fromEnd int
	if indices["WHERE"] != -1 {
		fromEnd = indices["WHERE"]
	} else if indices["ORDER BY"] != -1 {
		fromEnd = indices["ORDER BY"]
	} else if indices["LIMIT"] != -1 {
		fromEnd = indices["LIMIT"]
	} else {
		fromEnd = len(query)
	}

	fromPart := strings.TrimSpace(query[indices["FROM"]+5 : fromEnd])
	resources := strings.Split(fromPart, ",")
	if len(resources) != 1 {
		return fmt.Errorf("only one resource allowed in FROM clause")
	}

	allResources := []string{resources[0]}

	if err := o.parseResources(allResources, queryType); err != nil {
		return err
	}

	// Parse SELECT fields
	selectFields := strings.TrimSpace(query[6:indices["FROM"]])
	if err := o.parseFields(selectFields); err != nil {
		return err
	}

	// Parse WHERE clause if present
	if indices["WHERE"] != -1 {
		whereStart := indices["WHERE"] + 6
		var whereEnd int
		if indices["ORDER BY"] != -1 {
			whereEnd = indices["ORDER BY"]
		} else if indices["LIMIT"] != -1 {
			whereEnd = indices["LIMIT"]
		} else {
			whereEnd = len(query)
		}
		wherePart := strings.TrimSpace(query[whereStart:whereEnd])
		if wherePart == "" {
			return fmt.Errorf("WHERE clause cannot be empty")
		}
		o.requestedQuery = wherePart
	}

	// Parse ORDER BY clause if present
	if err := o.parseOrderBy(query, indices); err != nil {
		return err
	}

	// Parse LIMIT clause if present
	if err := o.parseLimit(query, indices); err != nil {
		return err
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
