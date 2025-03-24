/*
Copyright 2020 Yaacov Zamir <kobi.zamir@gmail.com>
and other contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Author: 2020 Yaacov Zamir <kobi.zamir@gmail.com>
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/yaacov/kubectl-sql/pkg/printers"
)

// SQLOptions provides information required to update
// the current context on a user's KUBECONFIG
type SQLOptions struct {
	configFlags *genericclioptions.ConfigFlags

	rawConfig              clientcmd.ClientConfig
	defaultSQLConfigPath   string
	requestedSQLConfigPath string
	args                   []string

	defaultAliases     map[string]string
	defaultTableFields printers.TableFieldsMap
	orderByFields      []printers.OrderByField
	limit              int

	namespace          string
	allNamespaces      bool
	requestedResources []string
	requestedQuery     string
	requestedOnQuery   string

	outputFormat string

	genericclioptions.IOStreams
}

// SQLConfig describes configuration overrides for SQL queries.
type SQLConfig struct {
	Aliases       map[string]string       `json:"aliases"`
	TableFields   printers.TableFieldsMap `json:"table-fields"`
	OrderByFields []printers.OrderByField `json:"order-by-fields,omitempty"`
	Limit         int                     `json:"limit,omitempty"`
}

// NewSQLConfig provides an instance of SQLConfig with default values
func NewSQLConfig() *SQLConfig {
	return &SQLConfig{
		Aliases:       defaultAliases,
		TableFields:   defaultTableFields,
		OrderByFields: []printers.OrderByField{},
		Limit:         0, // Default to no limit
	}
}

// Read SQL json config file.
func (o *SQLOptions) readConfigFile(filename string) error {
	userConfig := NewSQLConfig()

	// Init default config.
	o.defaultAliases = userConfig.Aliases
	o.defaultTableFields = userConfig.TableFields
	o.orderByFields = userConfig.OrderByFields
	o.limit = userConfig.Limit

	// If file is missing, fail quietly.
	if !fileExists(o.requestedSQLConfigPath) {
		return nil
	}

	file, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("can't read file '%s', %v", filename, err)
	}

	err = json.Unmarshal(file, &userConfig)
	if err != nil {
		return fmt.Errorf("can't parse json file '%s', %v", filename, err)
	}

	// Merge user defined aliases into the default aliases map.
	if len(userConfig.Aliases) > 0 {
		for k, v := range userConfig.Aliases {
			o.defaultAliases[k] = v
		}
	}

	// Merge user defined tables into the default table headers.
	if len(userConfig.TableFields) > 0 {
		for k, v := range userConfig.TableFields {
			o.defaultTableFields[k] = v
		}
	}

	// Set OrderByFields from config if specified
	if len(userConfig.OrderByFields) > 0 {
		o.orderByFields = userConfig.OrderByFields
	}

	// Set limit from config if specified
	if userConfig.Limit > 0 {
		o.limit = userConfig.Limit
	}

	return nil
}

// fileExists checks if a file exists and is not a directory.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
