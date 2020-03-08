
<p align="center">
  <img src="https://raw.githubusercontent.com/yaacov/kubectl-sql/master/img/kubesql-162.png" alt="kubectl-sql Logo">
</p>

# kubectl-sql

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

...  json
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
kubectl-sql join nodes,pods on \
    "nodes.status.addresses.1.address = pods.status.hostIP and not pods.phase ~= 'Running'" -A
...
```

#### Print help

``` bash
kubectl-sql --help
...
```
