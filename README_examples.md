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
kubectl-sql --all-namespaces "select * from pods where name ~= 'cni'"
NAMESPACE    	NAME               	PHASE  	hostIP        	CREATION_TIME(RFC3339)       	
openshift-cnv	ovs-cni-amd64-5vgcg	Running	192.168.126.58	2020-02-10T23:26:31+02:00    	
openshift-cnv	ovs-cni-amd64-8ts4w	Running	192.168.126.12	2020-02-10T22:01:59+02:00    	
openshift-cnv	ovs-cni-amd64-d6vdb	Running	192.168.126.53	2020-02-10T23:13:45+02:00
...
```

#### Namespaced

``` bash
# Get pods in namespace "openshift-multus" that hase name containing "ovs"
kubectl-sql "select * from openshift-multus/pods where name ~= 'cni'"
KIND: Pod	COUNT: 3
NAMESPACE       	NAME                               	PHASE  	CREATION_TIME(RFC3339)       	
openshift-multus	multus-additional-cni-plugins-7kcsd	Running	2024-12-02T11:41:45Z         	
openshift-multus	multus-additional-cni-plugins-kc8sz	Running	2024-12-02T11:41:45Z         	
openshift-multus	multus-additional-cni-plugins-vrpx9	Running	2024-12-02T11:41:45Z  
...
```

#### Select fields

``` bash
# Get pods in namespace "openshift-multus" that hase name containing "ovs"
# Select the fields name, status.phase, status.podIP
kubectl-sql "select name, status.phase, status.podIP from openshift-multus/pods where name ~= 'cni'"
KIND: Pod	COUNT: 3
name                               	status.phase	status.podIP	
multus-additional-cni-plugins-7kcsd	Running     	10.130.10.85	
multus-additional-cni-plugins-kc8sz	Running     	10.131.6.65 	
multus-additional-cni-plugins-vrpx9	Running     	10.129.8.252
...
```

#### Alias selected fields

``` bash
# Get pods in namespace "openshift-multus" that hase name containing "ovs"
# Select the fields name, status.phase as phase, status.podIP as ip
kubectl-sql "select name, status.phase as phase, status.podIP as ip \
  from openshift-multus/pods \
  where name ~= 'cni' and ip ~= '5$' and phase = 'Running'"
KIND: Pod	COUNT: 2
name                               	phase  	ip          	
multus-additional-cni-plugins-7kcsd	Running	10.130.10.85	
multus-additional-cni-plugins-kc8sz	Running	10.131.6.65 
...
```

#### Using Regexp

``` bash
# Get all pods from current namespace scope, that has a name starting with "virt-" and
# IP that ends with ".84"
kubectl-sql -n default "select * from pods where name ~= '^virt-' and status.podIP ~= '[.]84$'"
NAMESPACE	NAME                          	PHASE  	hostIP        	CREATION_TIME(RFC3339)       	
default  	virt-launcher-test-bdw2p-lcrwx	Running	192.168.126.56	2020-02-12T14:14:01+02:00
...
```

#### SI Units

``` bash
# Get all persistant volume clames that are less then 20Gi, and output as json.
kubectl-sql --all-namespaces -o json "select * from pvc where spec.resources.requests.storage < 20Gi"

...  json
{
  "storage": "10Gi"
}
...
```

#### Comparing fields

``` bash
# Get replicas sets with 3 replicas but less ready relicas
kubectl-sql --all-namespaces "select * from rs where spec.replicas = 3 and status.readyReplicas < spec.replicas"

...
```

#### Join

<p align="center">
   <a href="https://asciinema.org/a/AiBPT3SL7R9MgHCJV1tI0k6fU" target="_blank"><img src="https://asciinema.org/a/AiBPT3SL7R9MgHCJV1tI0k6fU.svg" /></a>
<p>
  
``` bash
# Display non running pods by nodes for all namespaces.
kubectl-sql "select nodes join pods on \
    nodes.status.addresses[1].address = pods.status.hostIP and not pods.phase ~= 'Running'" -A
...
```

#### Escaping Identifiers

``` bash
# Square brackets can be used for identifiers that include special characters.
./kubectl-sql --all-namespaces "select * from pods where name ~= 'cni' and metadata.labels[openshift.io/component] = 'network'"
...
```

#### Print help

``` bash
kubectl-sql --help
...
```
