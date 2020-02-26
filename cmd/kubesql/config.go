// Copyright 2020 Yaacov Zamir <kobi.zamir@gmail.com>
// and other contributors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Author: 2020 Yaacov Zamir <kobi.zamir@gmail.com>

// Package main.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

// User defined overides.
type config struct {
	Aliases     map[string]string `json:"aliases"`
	TableFields tableFieldsMap    `json:"table-fields"`
}

// tableField describes how to print a table column.
type tableField struct {
	Title    string `json:"title"`
	Name     string `json:"name"`
	width    int
	template string
}

type tableFields []tableField
type tableFieldsMap map[string]tableFields

// Get user quary.
func userQuery(c *cli.Context) ([]string, string) {
	resourceNames := c.String("resource")
	query := c.String("query")

	// Parse command args
	if len(resourceNames) == 0 && c.Args().Len() == 1 {
		resourceNames = c.Args().Get(0)
	} else if len(query) == 0 && c.Args().Len() == 3 && c.Args().Get(1) == "where" {
		resourceNames = c.Args().Get(0)
		query = c.Args().Get(2)
	}

	if len(resourceNames) == 0 {
		errExit("Failed to parse resource query", fmt.Errorf("missing resource name or query"))
	}

	return strings.Split(resourceNames, ","), query
}

// Read config file
func readConfigFile(s string) {
	file, err := ioutil.ReadFile(s)
	errExit("Failed to read config file %v\n", err)

	userConfig := config{}

	err = json.Unmarshal([]byte(file), &userConfig)
	errExit("Failed to unmarshal json config file %v\n", err)

	// Merge user defined aliases into the default aliases map.
	if len(userConfig.Aliases) > 0 {
		for k, v := range userConfig.Aliases {
			defaultAliases[k] = v
		}
	}

	// Merge user defined tables into the default table headers.
	if len(userConfig.TableFields) > 0 {
		for k, v := range userConfig.TableFields {
			defaultTableFields[k] = v
		}
	}
}

// Try to read user config file
func readKubeSQLConfigFile(c *cli.Context) {
	verbose := c.Bool("verbose")

	if c.String("config") != "" {
		// Try to read user defined config filename:
		debugLog(verbose, "Reading config file %s\n", c.String("config"))
		readConfigFile(c.String("config"))
	} else {
		// Try the default config filename:
		if home, err := os.UserHomeDir(); err == nil {
			config := fmt.Sprintf(defaultKubeSQLConfigPath, home)

			if fileExists(config) {
				debugLog(verbose, "Reading default config file %s\n", config)
				readConfigFile(config)
			}
		}
	}
}

// fileExists checks if a file exists and is not a directory.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
