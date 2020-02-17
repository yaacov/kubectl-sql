
[![Go Report Card](https://goreportcard.com/badge/github.com/yaacov/kubesql)](https://goreportcard.com/report/github.com/yaacov/kubesql)
[![Build Status](https://travis-ci.org/yaacov/kubesql.svg?branch=master)](https://travis-ci.org/yaacov/kubesql)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

<p align="center">
  <img src="https://raw.githubusercontent.com/yaacov/kubesql/master/img/kubesql-162.png" alt="kubesql Logo">
</p>

# kubesql

Use sql like language to query the Kubernetes cluster manager

## Install

``` bash
git clone git@github.com:yaacov/kubesql.git
cd kubesql

go build -o kubesql cmd/kubesql/*.go
```

## What doed it do

kubesql let you select Kubernetes resources based on the value of one or more resource fields, using
human readable easy to use SQL like query langauge.

For usage:
```
./kubesql --help
```

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

## Alternatives

### jq

`jq` is a lightweight and flexible command-line JSON processor. It is posible to
pipe the kubectl command output into the `jq` command to create complicted searches.

https://stedolan.github.io/jq/manual/#select(boolean_expression)

### kubectl --field-selector

Field selectors let you select Kubernetes resources based on the value of one or more resource fields. Here are some examples of field selector queries.

https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/
