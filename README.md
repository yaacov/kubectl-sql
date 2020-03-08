
[![Go Report Card](https://goreportcard.com/badge/github.com/yaacov/kubectl-sql)](https://goreportcard.com/report/github.com/yaacov/kubectl-sql)
[![Build Status](https://travis-ci.org/yaacov/kubectl-sql.svg?branch=master)](https://travis-ci.org/yaacov/kubectl-sql)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

<p align="center">
  <img src="https://raw.githubusercontent.com/yaacov/kubectl-sql/master/img/kubesql-162.png" alt="kubectl-sql Logo">
</p>

# kubectl-sql

kubectl-sql is a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/) that use SQL like language to query the [Kubernetes](https://kubernetes.io/) cluster manager

  - [More docs](#more-docs)
  - [Install](#install)
  - [What does it do ?](#what-does-it-do-)
  - [Examples](#examples)
  - [Alternatives](#alternatives)
    - [jq](#jq)
    - [Field Selector](#kubectl---field-selector)

## More docs

 - [The SQL like language](https://github.com/yaacov/kubectl-sql/blob/master/README_language.md)
 - [More examples](https://github.com/yaacov/kubectl-sql/blob/master/README_config.md)
 - [Using the config file](https://github.com/yaacov/kubectl-sql/blob/master/README_config.md)

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

## Examples

<p align="center">
   <a href="https://asciinema.org/a/vOSwHzeOLbVhQb79ajFmql2uk" target="_blank"><img src="https://asciinema.org/a/vOSwHzeOLbVhQb79ajFmql2uk.svg" /></a>
<p>

``` bash
# Get all pods from current namespace scope, that has a name starting with "virt-" and
# IP that ends with ".84"
kubectl-sql get pods where "name ~= '^virt-' and status.podIP ~= '[.]84$'"
AMESPACE	NAME                          	PHASE  	hostIP        	CREATION_TIME(RFC3339)       	
default  	virt-launcher-test-bdw2p-lcrwx	Running	192.168.126.56	2020-02-12T14:14:01+02:00
...
```

``` bash
# Get all persistant volume clames that are less then 20Gi, and output as json.
kubectl-sql -o json get pvc where "spec.resources.requests.storage < 20Gi"
...
```
  
``` bash
# Display non running pods by nodes for all namespaces.
kubectl-sql join nodes,pods on \
    "nodes.status.addresses.1.address = pods.status.hostIP and not pods.phase ~= 'Running'" -A
...
```

## Alternatives

#### jq

`jq` is a lightweight and flexible command-line JSON processor. It is possible to
pipe the kubectl command output into the `jq` command to create complicated searches ( [Illustrated jq toturial](https://github.com/MoserMichael/jq-illustrated) )

https://stedolan.github.io/jq/manual/#select(boolean_expression)

#### kubectl --field-selector

Field selectors let you select Kubernetes resources based on the value of one or more resource fields. Here are some examples of field selector queries.

https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/
