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

// Template for help message ( -h --help flages)
func helpTemplate() string {
	return `{{.Name}} - {{.Usage}}

Usage:
  {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}

Examples:
  # Query pods with name that matches /^test-.+/ ( e.g. name starts with "test-" )
  {{.HelpName}} get pods where "name ~= '^test-.+'"

  # Query replicasets where spec replicas is 3 or 5 and ready replicas is less then 3
  {{.HelpName}} get rs where "(spec.replicas = 3 or spec.replicas = 5) and status.readyReplicas < 3"

  # Query virtual machine instanses that are missing the lable "flavor.template.kubevirt.io/medium" 
  {{.HelpName}} get vmis where "labels.flavor.template.kubevirt.io/medium is null"

Special fields:
  name -> metadata.name
  namespace -> metadata.namespace
  labels -> metadata.labels
  created -> creation timestamp (RFC3339)
  deleted -> deletion timestamp (RFC3339)
  annotations -> metadata.annotations

Website:
   https://github.com/yaacov/kubesql
{{if len .Authors}}
Commands:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .VisibleFlags}}
Options:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}
Author:
   Yaacov Zamir

Copyright:
   Apache License
   Version 2.0, January 2004
   http://www.apache.org/licenses/
`
}

// Template for version message, no server version.
func versionTemplate() string {
	return "Client Version: %s\n"
}

// Template for version message, with server version.
func fullVersionTemplate() string {
	return "Client Version: %s\nServer Version: %s\n"
}
