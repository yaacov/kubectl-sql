package query

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var selectRegexp = regexp.MustCompile(`(?i)^(?:(sum|len|any|all)\s*\(?\s*([^)\s]+)\s*\)?|(.+?))\s*(?:as\s+(.+))?$`)

// Kubernetes naming validation
var k8sNameRegexp = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
var k8sNameMaxLength = 63

// Valid alias regex - must be a simple identifier without special characters
var validAliasRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// parseSelectClause splits and parses a select clause into SelectOptions entries.
func parseSelectClause(selectClause string) []SelectOption {
	var opts []SelectOption
	for _, raw := range strings.Split(selectClause, ",") {
		field := strings.TrimSpace(raw)
		if field == "" {
			continue
		}
		if m := selectRegexp.FindStringSubmatch(field); m != nil {
			reducer := strings.ToLower(m[1])
			expr := m[2]
			if expr == "" {
				expr = m[3]
			}

			// Clean up the expression from brackets and parentheses
			expr = strings.Trim(expr, ".()")

			alias := m[4]
			if alias == "" {
				// If no alias is provided, we will use the field name as the alias
				alias = ""
			}
			if !strings.HasPrefix(expr, ".") && !strings.HasPrefix(expr, "{") {
				expr = "." + expr
			}
			opts = append(opts, SelectOption{
				Field:   expr,
				Alias:   alias,
				Reducer: reducer,
			})
		}
	}
	return opts
}

// parseOrderByClause splits an ORDER BY clause into OrderOption entries.
func parseOrderByClause(orderByClause string, selectOpts []SelectOption) []OrderOption {
	var orderOpts []OrderOption

	for _, rawField := range strings.Split(orderByClause, ",") {
		fieldStr := strings.TrimSpace(rawField)
		if fieldStr == "" {
			continue
		}

		// determine direction
		parts := strings.Fields(fieldStr)
		descending := false
		last := parts[len(parts)-1]
		if strings.EqualFold(last, "desc") {
			descending = true
			parts = parts[:len(parts)-1]
		} else if strings.EqualFold(last, "asc") {
			parts = parts[:len(parts)-1]
		}

		// ensure JSONPath format
		name := strings.Join(parts, " ")
		if !strings.HasPrefix(name, ".") && !strings.HasPrefix(name, "{") {
			name = "." + name
		}

		// find matching select option or create default
		var selOpt SelectOption
		found := false
		for _, sel := range selectOpts {
			if sel.Field == name || sel.Alias == strings.TrimPrefix(name, ".") {
				selOpt = sel
				found = true
				break
			}
		}
		if !found {
			selOpt = SelectOption{
				Field:   name,
				Alias:   strings.TrimPrefix(name, "."),
				Reducer: "",
			}
		}

		orderOpts = append(orderOpts, OrderOption{
			Field:      selOpt,
			Descending: descending,
		})
	}

	return orderOpts
}

// validateJSONPath checks if a field path is valid for GetValueByPathString
func validateJSONPath(path string) error {
	// Clean up path similar to what GetValueByPathString does
	path = strings.TrimPrefix(path, "{{")
	path = strings.TrimSuffix(path, "}}")
	path = strings.TrimSpace(path)

	// Remove leading dot as GetValueByPathString does
	path = strings.TrimPrefix(path, ".")

	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check for balanced brackets
	bracketCount := 0
	for _, char := range path {
		if char == '[' {
			bracketCount++
		} else if char == ']' {
			bracketCount--
			if bracketCount < 0 {
				return fmt.Errorf("unbalanced brackets in path: %s", path)
			}
		}
	}

	if bracketCount != 0 {
		return fmt.Errorf("unbalanced brackets in path: %s", path)
	}

	// Check for invalid characters
	for _, segment := range strings.Split(path, ".") {
		// Skip empty segments (would happen with double dots)
		if segment == "" {
			return fmt.Errorf("invalid empty segment in path: %s", path)
		}

		// Remove array notation for this check
		baseName := segment
		bracketIdx := strings.Index(baseName, "[")
		if bracketIdx > 0 {
			baseName = baseName[:bracketIdx]
		}

		// Check for spaces and special characters in the base name
		if strings.ContainsAny(baseName, " \t\n\r{}()'+*/%&|^!=<>,;:`\\\"") {
			return fmt.Errorf("segment '%s' contains invalid characters in path: %s", baseName, path)
		}
	}

	return nil
}

// validateK8sName validates that a string is a valid Kubernetes resource name
func validateK8sName(name, resourceType string) error {
	if name == "" {
		return fmt.Errorf("%s name cannot be empty", resourceType)
	}

	if len(name) > k8sNameMaxLength {
		return fmt.Errorf("%s name '%s' is too long (maximum %d characters)", resourceType, name, k8sNameMaxLength)
	}

	if !k8sNameRegexp.MatchString(name) {
		return fmt.Errorf("%s name '%s' is invalid: must consist of lowercase alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character", resourceType, name)
	}

	return nil
}

// validateSelectOptions validates each SelectOption in a slice of options
func validateSelectOptions(selectOpts []SelectOption) error {
	if len(selectOpts) == 0 {
		// Empty select is acceptable
		return nil
	}

	// Check for "*" for default selection
	if len(selectOpts) == 1 && selectOpts[0].Field == ".*" {
		return nil
	}

	for i, opt := range selectOpts {
		// Validate Field
		if opt.Field == "" {
			return fmt.Errorf("select field at position %d cannot be empty", i+1)
		}

		// Validate field is a valid JSONPath for GetValueByPathString
		if err := validateJSONPath(opt.Field); err != nil {
			return fmt.Errorf("invalid field at position %d: %v", i+1, err)
		}

		// Validate Alias if present
		if opt.Alias != "" {
			// Check for spaces and special characters
			if !validAliasRegexp.MatchString(opt.Alias) {
				return fmt.Errorf("invalid alias '%s': must contain only alphanumeric characters and underscores", opt.Alias)
			}
		}

		// Validate Reducer if present
		if opt.Reducer != "" {
			switch opt.Reducer {
			case "sum", "len", "any", "all":
				// Valid reducers
			default:
				return fmt.Errorf("invalid reducer '%s' at position %d: must be one of 'sum', 'len', 'any', 'all'", opt.Reducer, i+1)
			}
		}
	}

	return nil
}

// validateFromOptions validates the name and namespace in FromOptions
func validateFromOptions(from FromOptions) error {
	// Resource name must not be empty
	if from.Name == "" {
		return fmt.Errorf("resource name cannot be empty")
	}

	// Validate resource name against K8s naming rules
	if err := validateK8sName(from.Name, "resource"); err != nil {
		return err
	}

	// For namespace: can be empty, "*", or valid K8s name
	if from.Namespace != "" && from.Namespace != "*" {
		if err := validateK8sName(from.Namespace, "namespace"); err != nil {
			return err
		}
	}

	return nil
}

// ParseQueryString parses a query string into its component parts
func ParseQueryString(query string) (*QueryOptions, error) {
	options := &QueryOptions{
		Limit: -1, // Default to no limit
	}

	if query == "" {
		return options, nil
	}

	// Convert query to lowercase for case-insensitive matching but preserve original for extraction
	queryLower := strings.ToLower(query)

	// Check for all clause positions
	selectIndex := strings.Index(queryLower, "select ")
	fromIndex := strings.Index(queryLower, "from ")
	whereIndex := strings.Index(queryLower, "where ")
	orderByIndex := strings.Index(queryLower, "order by ")
	limitIndex := strings.Index(queryLower, "limit ")

	// Extract SELECT clause if it exists
	if selectIndex >= 0 {
		selectEnd := len(query)
		if fromIndex > selectIndex {
			selectEnd = fromIndex
		} else if whereIndex > selectIndex {
			selectEnd = whereIndex
		} else if orderByIndex > selectIndex {
			selectEnd = orderByIndex
		} else if limitIndex > selectIndex {
			selectEnd = limitIndex
		}

		// Extract select clause (skip "select " prefix which is 7 chars)
		selectClause := strings.TrimSpace(query[selectIndex+7 : selectEnd])
		options.Select = parseSelectClause(selectClause)
		options.HasSelect = len(options.Select) > 0

		// Validate the parsed SELECT options
		if err := validateSelectOptions(options.Select); err != nil {
			return nil, err
		}
	}

	// Extract FROM clause if it exists
	if fromIndex >= 0 {
		fromEnd := len(query)
		if whereIndex > fromIndex {
			fromEnd = whereIndex
		} else if orderByIndex > fromIndex {
			fromEnd = orderByIndex
		} else if limitIndex > fromIndex {
			fromEnd = limitIndex
		}

		// Extract from clause (skip "from " prefix which is 5 chars)
		fromClause := strings.TrimSpace(query[fromIndex+5 : fromEnd])

		// Parse the from clause without validation
		if fromClause == "" {
			return nil, fmt.Errorf("FROM clause requires a resource name")
		}

		options.HasFrom = true

		// Handle the different FROM clause formats
		if strings.Contains(fromClause, "/") {
			parts := strings.SplitN(fromClause, "/", 2)
			if parts[0] == "*" {
				// Format: "from */name"
				options.From.Name = parts[1]
				options.From.AllNamespaces = true
			} else {
				// Format: "from namespace/name"
				options.From.Namespace = parts[0]
				options.From.Name = parts[1]
			}
		} else {
			// Format: "from name"
			options.From.Name = fromClause
		}

		// Validate the parsed FROM options
		if err := validateFromOptions(options.From); err != nil {
			return nil, err
		}
	}

	// Extract WHERE clause if it exists
	if whereIndex >= 0 {
		whereEnd := len(query)
		if orderByIndex > whereIndex {
			whereEnd = orderByIndex
		} else if limitIndex > whereIndex {
			whereEnd = limitIndex
		}

		// Extract where clause (skip "where " prefix which is 6 chars)
		options.Where = strings.TrimSpace(query[whereIndex+6 : whereEnd])
	}

	// Extract ORDER BY clause if it exists
	if orderByIndex >= 0 {
		orderByEnd := len(query)
		if limitIndex > orderByIndex {
			orderByEnd = limitIndex
		}

		// Extract order by clause (skip "order by " prefix which is 9 chars)
		orderByClause := strings.TrimSpace(query[orderByIndex+9 : orderByEnd])

		// use helper to build OrderOption slice
		options.OrderBy = parseOrderByClause(orderByClause, options.Select)
		options.HasOrderBy = len(options.OrderBy) > 0
	}

	// Extract LIMIT clause using regex (for simplicity with number extraction)
	limitRegex := regexp.MustCompile(`(?i)limit\s+(\d+)`)
	limitMatches := limitRegex.FindStringSubmatch(query)
	if len(limitMatches) > 1 {
		limit, err := strconv.Atoi(limitMatches[1])
		if err != nil {
			return nil, fmt.Errorf("invalid limit value: %v", err)
		}
		options.Limit = limit
		options.HasLimit = true
	}

	return options, nil
}
