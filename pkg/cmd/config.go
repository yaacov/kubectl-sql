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
	"io/ioutil"
	"os"

	"github.com/yaacov/kubectl-sql/pkg/printers"
)

// SQLConfig describes configuration overides for SQL queries.
type SQLConfig struct {
	Aliases     map[string]string       `json:"aliases"`
	TableFields printers.TableFieldsMap `json:"table-fields"`
}

// NewSQLConfig provides an instance of SQLConfig with default values
func NewSQLConfig() *SQLConfig {
	return &SQLConfig{
		Aliases: map[string]string{
			"phase": "status.phase",
		},
		TableFields: printers.TableFieldsMap{
			"other": {
				{
					Title: "NAMESPACE",
					Name:  "namespace",
				},
				{
					Title: "NAME",
					Name:  "name",
				},
				{
					Title: "PHASE",
					Name:  "status.phase",
				},
				{
					Title: "CREATION_TIME(RFC3339)",
					Name:  "created",
				},
			},
		},
	}
}

// Read SQL json config file.
func (o *SQLOptions) readConfigFile(filename string) error {
	userConfig := NewSQLConfig()

	// Init default config.
	o.defaultAliases = userConfig.Aliases
	o.defaultTableFields = userConfig.TableFields

	// If file is missing, fail quietly.
	if !fileExists(o.requestedSQLConfigPath) {
		return nil
	}

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("can't read file '%s', %v", filename, err)
	}

	err = json.Unmarshal([]byte(file), &userConfig)
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
