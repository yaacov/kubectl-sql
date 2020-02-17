
[![Go Report Card](https://goreportcard.com/badge/github.com/yaacov/kubesql)](https://goreportcard.com/report/github.com/yaacov/kubesql)
[![Build Status](https://travis-ci.org/yaacov/kubesql.svg?branch=master)](https://travis-ci.org/yaacov/kubesql)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

<p align="center">
  <img src="https://raw.githubusercontent.com/yaacov/kubesql/master/img/kubesql-162.png" alt="kubesql Logo">
</p>

# kubesql

Use sql like language to query the Kubernetes cluster manager

  - [Install](#install)
  - [What does it do ?](#what-does-it-do-)
  - [Examples](#examples)
  - [Query language](#query-language)
  - [Alternatives](#alternatives)

## Install

From source:

``` bash
git clone git@github.com:yaacov/kubesql.git
cd kubesql

go build -o kubesql cmd/kubesql/*.go
```

## What does it do ?

kubesql provides a simple and easy to use way to search for Kubernetes resources.

kubesql let you select Kubernetes resources based on the value of one or more resource fields, using
human readable easy to use SQL like query langauge.

Example:
``` bash
# Filter pods belonging to namespaces that start with "test-"
kubesql get pods where "namespace ~= '^test-'"
```

For other ways to select Kubernetes resources see [#Alternatives](https://github.com/yaacov/kubesql#alternatives).

#### Available Operators:

  - `=` : Equal
  - `~=` : Match Regular expression
  - `!=`, `!~` : Not Equal, Not matching Regular expression
  - `>`, `<`, `<=` and `>=` : Compere operators for strings and numbers
  - `is null`, `is not null` : Check field existance
  - `or`, `and`, `not` and `( )`

#### Aliases:
  - name -> metadata.name
  - namespace -> metadata.namespace
  - labels -> metadata.labels
  - creation -> creation timestamp
  - deletion -> deletion timestamp
  - annotations -> metadata.annotations

#### Output formats:
  - Table
  - YAML
  - JSON

#### Arrays and lists:
kubesql does not support list fields.

## Examples

``` bash
# Get pods that hase name containing "ovs"
./kubesql --all-namespaces get pods where "name ~= 'ovs'"

openshift-cnv                  ovs-cni-amd64-5vgcg            2020-02-10T23:26:31+02:00
openshift-cnv                  ovs-cni-amd64-8ts4w            2020-02-10T22:01:59+02:00
openshift-cnv                  ovs-cni-amd64-d6vdb            2020-02-10T23:13:45+02:00
openshift-cnv                  ovs-cni-amd64-gxvm4            2020-02-10T22:01:59+02:00
...
```

``` bash
# Get all pods from current namespace scope, that has a name starting with "virt-" and
# IP that ends with ".84"
./kubesql get pods where "name ~= '^virt-' and status.podIP ~= '[.]84$'"
default                        virt-launcher-test-bdw2p-lcrwx 2020-02-12T14:14:01+02:00
...
```

``` bash
# Get all persistant volume clames that are less then 20Gi, and output as json.
./kubesql -o json get pvc where "spec.resources.requests.storage ~= '^1[0-9]Gi' or spec.resources.requests.storage ~= '^[1-9]Gi'" | jq .Object.spec.resources.requests

{
  "storage": "10Gi"
}
...
```

``` bash
# Get replicas sets with 3 replicas but less ready relicas
./kubesql -A -o yaml get rs where "spec.replicas = 3 and status.readyReplicas < 3"

...
```
### Print help

```
./kubesql --help

kubesql - uses sql like language to query the Kubernetes cluster manager.

Usage:
  kubesql [global options] command [command options] [arguments...]

Examples:
  # Query pods with name that matches /^test-.+/ ( e.g. name starts with "test-" )
  kubesql get pods where "name ~= '^test-.+'"

  # Query replicasets where spec replicas is 3 or 5 and ready replicas is less then 3
  kubesql get rs where "(spec.replicas = 3 or spec.replicas = 5) and status.readyReplicas < 3"

  # Query virtual machine instanses that are missing the lable "flavor.template.kubevirt.io/medium" 
  kubesql get vmis where "labels.flavor.template.kubevirt.io/medium is null"

Special fields:
  name -> metadata.name
  namespace -> metadata.namespace
  labels -> metadata.labels
  creation -> creation timestamp
  deletion -> deletion timestamp
  annotations -> metadata.annotations

Website:
   https://github.com/yaacov/kubesql

Options:
   --kubeconfig value           Path to the kubeconfig file to use for CLI requests.
   --namespace value, -n value  If present, the namespace scope for this CLI request.
   --output value, -o value     Output format, options: table, yaml or json. (default: "table")
   --all-namespaces, -A         Use all namespace scopes for this CLI request. (default: false)
   --verbose, -V                Show verbose output (default: false)
   --help, -h                   show help (default: false)
   --version, -v                print the version (default: false)
   
Author:
   Yaacov Zamir

Copyright:
   Apache License
   Version 2.0, January 2004
   http://www.apache.org/licenses/

```

## Query language

qubesql uses Tree Search Language (TSL). TSL is a wonderful human readable filtering language.

https://github.com/yaacov/tree-search-language

## Alternatives

### jq

`jq` is a lightweight and flexible command-line JSON processor. It is posible to
pipe the kubectl command output into the `jq` command to create complicted searches.

https://stedolan.github.io/jq/manual/#select(boolean_expression)

### kubectl --field-selector

Field selectors let you select Kubernetes resources based on the value of one or more resource fields. Here are some examples of field selector queries.

https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/
