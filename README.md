
[![Go Report Card](https://goreportcard.com/badge/github.com/yaacov/kubectl-sql)](https://goreportcard.com/report/github.com/yaacov/kubectl-sql)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

<p align="center">
  <img src="https://raw.githubusercontent.com/yaacov/kubectl-sql/master/img/kubesql-248.png" alt="kubectl-sql Logo">
</p>

# kubectl-sql

kubectl-sql is a [kubectl plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/) that use SQL like language to query the [Kubernetes](https://kubernetes.io/) cluster manager

  - [Install](#install)
  - [What can I do with it ?](#what-can-i-do-with-it-)
  - [Alternatives](#alternatives)

<p align="center">
  <a href="https://asciinema.org/a/308607" target="_blank"><img src="https://asciinema.org/a/308607.svg" /></a>
<p>

## More docs

 - [kubectl-sql's query language](https://github.com/yaacov/kubectl-sql/blob/master/README_language.md)
 - [More kubectl-sql examples](https://github.com/yaacov/kubectl-sql/blob/master/README_examples.md)
 - [Using the config file](https://github.com/yaacov/kubectl-sql/blob/master/README_config.md)

## Install

Using [krew](https://sigs.k8s.io/krew) plugin manager to install:

```
kubectl krew install sql
kubectl sql --help
```

Using Fedora Copr:

``` bash
dnf copr enable yaacov/kubesql
dnf install kubectl-sql
```

From source:

``` bash
# Clone code
git clone git@github.com:yaacov/kubectl-sql.git
cd kubectl-sql

# Build kubectl-sql
make

# Install into local machine PATH
sudo install ./kubectl-sql /usr/local/bin/
```

<p align="center">
   <a href="https://asciinema.org/a/jPQQCjFG2qGqlZ6HKXWoQjFWa" target="_blank"><img src="https://asciinema.org/a/jPQQCjFG2qGqlZ6HKXWoQjFWa.svg" /></a>
<p>
 
## What can I do with it ?

kubectl-sql let you select Kubernetes resources based on the value of one or more resource fields, using
human readable easy to use SQL like query language. It is also posible to find connected resources useing the
`join` command.

[More kubectl-sql examples](https://github.com/yaacov/kubectl-sql/blob/master/README_examples.md)

``` bash
# Get all pods from current namespace scope, that has a name starting with "virt-" and
# IP that ends with ".84"
kubectl-sql "select * from pods where name ~= '^virt-' and status.podIP ~= '[.]84$'"
AMESPACE	NAME                          	PHASE  	hostIP        	CREATION_TIME(RFC3339)       	
default  	virt-launcher-test-bdw2p-lcrwx	Running	192.168.126.56	2020-02-12T14:14:01+02:00
...
```

``` bash
# Get all persistant volume clames that are less then 20Gi, and output as json.
kubectl-sql -o json "select * from pvc where spec.resources.requests.storage < 20Gi"
...
```
  
``` bash
# Display non running pods by nodes for all namespaces.
kubectl-sql "select * from nodes join pods on \
    nodes.status.addresses[1].address = pods.status.hostIP and not pods.phase ~= 'Running'" -A
...
```

``` bash
# Filter replica sets with less ready-replicas then replicas"
kubectl-sql --all-namespaces "select * from rs where status.readyReplicas < status.replicas"
```

<p align="center">
   <a href="https://asciinema.org/a/vOSwHzeOLbVhQb79ajFmql2uk" target="_blank"><img src="https://asciinema.org/a/vOSwHzeOLbVhQb79ajFmql2uk.svg" /></a>
<p>

<p align="center">
  <a href="https://asciinema.org/a/308443" target="_blank"><img src="https://asciinema.org/a/308443.svg" /></a>
<p>

<p align="center">
   <a href="https://asciinema.org/a/308434" target="_blank"><img src="https://asciinema.org/a/308434.svg" /></a>
<p>

#### Output formats:
| --output flag | Print format |
|----|---|
| table | Table |
| name | Names only |
| yaml | YAML |
| json | JSON |

## Alternatives

#### jq

`jq` is a lightweight and flexible command-line JSON processor. It is possible to
pipe the kubectl command output into the `jq` command to create complicated searches ( [Illustrated jq toturial](https://github.com/MoserMichael/jq-illustrated) )

https://stedolan.github.io/jq/manual/#select(boolean_expression)

#### kubectl --field-selector

Field selectors let you select Kubernetes resources based on the value of one or more resource fields. Here are some examples of field selector queries.

https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/
