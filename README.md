
[![Go Report Card](https://goreportcard.com/badge/github.com/yaacov/kubectl-sql)](https://goreportcard.com/report/github.com/yaacov/kubectl-sql)
[![Build Status](https://travis-ci.org/yaacov/kubectl-sql.svg?branch=master)](https://travis-ci.org/yaacov/kubectl-sql)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

<p align="center">
  <img src="https://raw.githubusercontent.com/yaacov/kubectl-sql/master/img/kubesql-162.png" alt="kubectl-sql Logo">
</p>

# kubectl-sql

kubectl-sql is a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/) that use SQL like language to query the [Kubernetes](https://kubernetes.io/) cluster manager

  - [Install](#install)
  - [What does it do ?](#what-does-it-do-)
    - [Operators](#available-operators)
    - [Math Operators](#available-math-operators)
    - [Aliases](#aliases)
    - [SI Units](#si-units)
    - [Formats](#output-formats)
    - [Arrays](#arrays-and-lists)
    - [Escaping](#identifier-escaping)
  - [Examples](#examples)
    - [All Namespaces](#all-namespaces)
    - [Regexp](#using-regexp)
    - [Compere Fields](#comparing-fields)
    - [Join](#join)
    - [Print Help](#print-help)
  - [Config File](#config-file)
  - [Query language](#query-language)
  - [Alternatives](#alternatives)
    - [jq](#jq)
    - [Field Selector](#kubectl---field-selector)
  
## Install

<p align="center">
   <a href="https://asciinema.org/a/jPQQCjFG2qGqlZ6HKXWoQjFWa" target="_blank"><img src="https://asciinema.org/a/jPQQCjFG2qGqlZ6HKXWoQjFWa.svg" /></a>
<p>

Using `go get` command:
``` bash
GO111MODULE=on go get -v github.com/yaacov/kubectl-sql/cmd/kubectl-sql
```

Using Fedora Copr:

```
dnf copr enable yaacov/kubesql
dnf install kubectl-sql
```

From source:

``` bash
git clone git@github.com:yaacov/kubectl-sql.git
cd kubectl-sql

make
```

#### Plugin

With `kubectl` v1.12 or higher, `kubectl-sql` can be used as a kubectl plugin.

``` bash
# Using kubectl-sql as a kubectl plugin.
kubectl sql get pods where "name ilike 'test-%'"
```

## What does it do ?

<p align="center">
  <a href="https://asciinema.org/a/308443" target="_blank"><img src="https://asciinema.org/a/308443.svg" /></a>
<p>

kubectl-sql let you select Kubernetes resources based on the value of one or more resource fields, using
human readable easy to use SQL like query language.

Example:
``` bash
# Filter replica sets with less ready-replicas then replicas"
kubectl-sql --all-namespaces get rs where "status.readyReplicas < status.replicas"
```

For other ways to select Kubernetes resources see [#Alternatives](https://github.com/yaacov/kubectl-sql#alternatives).

#### Available Operators:

| Operators | Example |
|----|---|
| `=`, `~=` | `name ~= '^test-'`  |
| `like`, `ilike` | `phase ilike 'run%'`  | 
|`!=`, `!~` |  `namespace != 'default'` |
|`>`, `<`, `<=` and `>=` | `created < 2020-01-15T00:00:00Z` |
|`is null`, `is not null`| `spec.domain.cpu.dedicatedCpuPlacement is not null` |
| `in`   |  `spec.domain.resources.limits.memory in (1Gi, 2Gi)` |
| `between`   |  `spec.domain.resources.limits.memory between 1Gi and 2Gi` |
| `or`, `and` and `not` | `name ~= 'virt-' and namespace != 'test-wegsb'` |
| `( )`|  `phase = 'Running' and (namespace ~= 'cnv-' or namespace ~= 'virt-')`|

#### Available Math Operators:

| Operators | Notes |
|----|---|
| `+`, `-` | Addition and Subtraction |
| `*`, `/` | Multiplication and Division |
| `( )`|  |

#### Aliases:
| Alias | Resource Path | Example |
|----|---|---|
| name | metadata.name | |
| namespace | metadata.namespace | `namespace ~= '^test-[a-z]+$'` |
| labels | metadata.labels | |
| annotations | metadata.annotations | |
| created | creation timestamp | |
| deleted | deletion timestamp | |
| phase | status.phase | `phase = 'Running'` |

#### SI Units:
| Unit | Multiplier | Example |
|----|---|---|
| Ki | 1024 | |
| Mi | 1024^2 | `spec.containers.1.resources.requests.memory = 200Mi` |
| Gi | 1024^3 | |
| Ti | 1024^4 | |
| Pi | 1024^5 | |

#### Booleans:
| Example |
|---|
| `status.conditions.1.status = true` |

#### RFC3339 dates:
| Example |
|---|
| `status.conditions.1.lastTransitionTime > 2020-02-20T11:12:38Z`  |

#### Output formats:
| --output flag | Print format |
|----|---|
| table | Table |
| name | Names only |
| yaml | YAML |
| json | JSON |

#### Arrays and lists:
kubectl-sql support resource paths that include lists by using the list index as a field key.

``` bash
# Get the memory request for the first container.
kubectl-sql --all-namespaces get pods where "spec.containers.1.resources.requests.memory = 200Mi"
```

#### Identifier escaping

If identifier include characters that need escaping ( e.g. "-" or ":") it is possible
to escape the identifier name by wrapping it with `[...]` , `` `...` `` or `"..."`

## Examples

<p align="center">
   <a href="https://asciinema.org/a/vOSwHzeOLbVhQb79ajFmql2uk" target="_blank"><img src="https://asciinema.org/a/vOSwHzeOLbVhQb79ajFmql2uk.svg" /></a>
<p>
  
#### All namespaces

``` bash
# Get pods that hase name containing "ovs"
kubectl-sql --all-namespaces get pods where "name ~= 'ovs'"
AMESPACE    	NAME               	PHASE  	hostIP        	CREATION_TIME(RFC3339)       	
openshift-cnv	ovs-cni-amd64-5vgcg	Running	192.168.126.58	2020-02-10T23:26:31+02:00    	
openshift-cnv	ovs-cni-amd64-8ts4w	Running	192.168.126.12	2020-02-10T22:01:59+02:00    	
openshift-cnv	ovs-cni-amd64-d6vdb	Running	192.168.126.53	2020-02-10T23:13:45+02:00
...
```

#### Using Regexp

``` bash
# Get all pods from current namespace scope, that has a name starting with "virt-" and
# IP that ends with ".84"
kubectl-sql get pods where "name ~= '^virt-' and status.podIP ~= '[.]84$'"
AMESPACE	NAME                          	PHASE  	hostIP        	CREATION_TIME(RFC3339)       	
default  	virt-launcher-test-bdw2p-lcrwx	Running	192.168.126.56	2020-02-12T14:14:01+02:00
...
```
#### SI Units

``` bash
# Get all persistant volume clames that are less then 20Gi, and output as json.
kubectl-sql -o json get pvc where "spec.resources.requests.storage < 20Gi"

... 
{
  "storage": "10Gi"
}
...
```

#### Comparing fields

``` bash
# Get replicas sets with 3 replicas but less ready relicas
kubectl-sql --all-namespaces get rs where "spec.replicas = 3 and status.readyReplicas < spec.replicas"

...
```

#### Join


<p align="center">
   <a href="https://asciinema.org/a/AiBPT3SL7R9MgHCJV1tI0k6fU" target="_blank"><img src="https://asciinema.org/a/AiBPT3SL7R9MgHCJV1tI0k6fU.svg" /></a>
<p>
  
``` bash
# Display non running pods by nodes for all namespaces.
kubectl-sql join nodes,pods on "nodes.status.addresses.1.address = pods.status.hostIP and not pods.phase ~= 'Running'" -A
...
```

#### Print help

```
kubectl-sql --help
...
```

## Config File

<p align="center">
   <a href="https://asciinema.org/a/308440" target="_blank"><img src="https://asciinema.org/a/308440.svg" /></a>
<p>
  
Users can add aliases and edit the fields displayed in table view using json config files,
[see the example config file](https://github.com/yaacov/kubectl-sql/blob/master/kubectl-sql.json).

Flag: `--kubectl-sql <config file path>` (default: `$HOME/.kube/kubectl-sql.json`)

Example:
``` bash
kubectl-sql --kubectl-sql ./kubectl-sql.json get pods
...
```

## Query language

kubectl-sql uses Tree Search Language (TSL). TSL is a wonderful human readable filtering language.

https://github.com/yaacov/tree-search-language

## Alternatives

#### jq

`jq` is a lightweight and flexible command-line JSON processor. It is possible to
pipe the kubectl command output into the `jq` command to create complicated searches.

https://stedolan.github.io/jq/manual/#select(boolean_expression)

#### kubectl --field-selector

Field selectors let you select Kubernetes resources based on the value of one or more resource fields. Here are some examples of field selector queries.

https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/
