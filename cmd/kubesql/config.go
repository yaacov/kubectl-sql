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
	"io/ioutil"
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
