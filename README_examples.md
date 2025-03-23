<p align="center">
  <img src="https://raw.githubusercontent.com/yaacov/kubectl-sql/master/img/kubesql-162.png" alt="kubectl-sql Logo">
</p>

# kubectl-sql

## Examples

<p align="center">
   <a href="https://asciinema.org/a/vOSwHzeOLbVhQb79ajFmql2uk" target="_blank"><img src="https://asciinema.org/a/vOSwHzeOLbVhQb79ajFmql2uk.svg" /></a>
<p>

---

**Basic Selection & Namespace Filtering**

* **Select all pods in `default`:**
    ```bash
    kubectl sql "SELECT * FROM default/pods"
    ```
* **Names & namespaces of deployments:**
    ```bash
    kubectl sql "SELECT name, namespace FROM */deployments"
    ```
* **Service names & types in `kube-system`:**
    ```bash
    kubectl sql "SELECT name, spec.type FROM kube-system/services"
    ```

---

**Filtering with `WHERE` Clause**

* **Pods with label `app=my-app`:**
    ```bash
    kubectl sql "SELECT name FROM */pods WHERE metadata.labels.app = 'my-app'"
    ```
* **Deployments with image `nginx.*`:**
    ```bash
    kubectl sql "SELECT name FROM */deployments WHERE spec.template.spec.containers[1].image ~= 'nginx.*'"
    ```
* **Services of type `LoadBalancer`:**
    ```bash
    kubectl sql "SELECT name FROM */services WHERE spec.type = 'LoadBalancer'"
    ```
* **Pods not `Running`:**
    ```bash
    kubectl sql "SELECT name, status.phase FROM */pods WHERE status.phase != 'Running'"
    ```
* **Pods with container named nginx:**
    ```bash
    kubectl sql "SELECT name from */pods where spec.containers[1].name = 'nginx'"
    ```

---

**Aliasing with `AS` Keyword**

* **Alias `status.phase` to `pod_phase`:**
    ```bash
    kubectl sql "SELECT name, status.phase AS pod_phase FROM */pods"
    ```
* **Alias container image to `container_image`:**
    ```bash
    kubectl sql "SELECT name, spec.template.spec.containers[1].image AS container_image FROM */deployments"
    ```

---

**Time-Based Filtering (using `date`)**

* **Pods created in last 24 hours:**
    ```bash
    kubectl sql "SELECT name, metadata.creationTimestamp FROM */pods WHERE metadata.creationTimestamp > '$(date -Iseconds -d "24 hours ago")'"
    ```
* **Events related to pods in last 10 minutes:**
    ```bash
    kubectl sql "SELECT message, metadata.creationTimestamp, involvedObject.name FROM */events WHERE involvedObject.kind = 'Pod' AND metadata.creationTimestamp > '$(date -Iseconds -d "10 minutes ago")'"
    ```

---

**SI Extension Filtering**

* **Deployments with memory request < 512Mi:**
    ```bash
    kubectl sql "SELECT name, spec.template.spec.containers[1].resources.requests.memory FROM */deployments WHERE spec.template.spec.containers[1].resources.requests.memory < 512Mi"
    ```
* **PVCs with storage request > 10Gi:**
    ```bash
    kubectl sql "SELECT name, spec.resources.requests.storage FROM */persistentvolumeclaims WHERE spec.resources.requests.storage > 10Gi"
    ```
* **Pods with container memory limit > 1Gi:**
    ```bash
    kubectl sql "SELECT name, spec.containers[1].resources.limits.memory FROM */pods WHERE spec.containers[1].resources.limits.memory > 1Gi"
    ```

---

**All namespaces**

* **Get pods that have name containing "ovs" using regular kubectl all namespaces arg:**
    ```bash
    kubectl-sql --all-namespaces "select * from pods where name ~= 'cni'"
    NAMESPACE    	NAME               	PHASE  	hostIP        	CREATION_TIME(RFC3339)       	
    openshift-cnv	ovs-cni-amd64-5vgcg	Running	192.168.126.58	2020-02-10T23:26:31+02:00    	
    openshift-cnv	ovs-cni-amd64-8ts4w	Running	192.168.126.12	2020-02-10T22:01:59+02:00    	
    openshift-cnv	ovs-cni-amd64-d6vdb	Running	192.168.126.53	2020-02-10T23:13:45+02:00
    ...
    ```

---

**Namespaced**

* **Get pods in namespace "openshift-multus" that have name containing "ovs":**
    ```bash
    kubectl-sql -n openshift-multus "select * from pods where name ~= 'cni'"
    KIND: Pod	COUNT: 3
    NAMESPACE       	NAME                               	PHASE  	CREATION_TIME(RFC3339)       	
    openshift-multus	multus-additional-cni-plugins-7kcsd	Running	2024-12-02T11:41:45Z         	
    openshift-multus	multus-additional-cni-plugins-kc8sz	Running	2024-12-02T11:41:45Z         	
    openshift-multus	multus-additional-cni-plugins-vrpx9	Running	2024-12-02T11:41:45Z  
    ...
    ```

---

**Select fields**

* **Get pods in namespace "openshift-multus" with name containing "cni" and select specific fields:**
    ```bash
    kubectl-sql "select name, status.phase, status.podIP \
      from openshift-multus/pods \
      where name ~= 'cni'"
    KIND: Pod	COUNT: 3
    name                               	status.phase	status.podIP	
    multus-additional-cni-plugins-7kcsd	Running     	10.130.10.85	
    multus-additional-cni-plugins-kc8sz	Running     	10.131.6.65 	
    multus-additional-cni-plugins-vrpx9	Running     	10.129.8.252
    ...
    ```

---

**Alias selected fields**

* **Get pods matching criteria with aliased fields:**
    ```bash
    kubectl-sql "select name, status.phase as phase, status.podIP as ip \
      from openshift-multus/pods \
      where name ~= 'cni' and ip ~= '5$' and phase = 'Running'"
    KIND: Pod	COUNT: 2
    name                               	phase  	ip          	
    multus-additional-cni-plugins-7kcsd	Running	10.130.10.85	
    multus-additional-cni-plugins-kc8sz	Running	10.131.6.65 
    ...
    ```

---

**Using Regexp**

* **Get pods with name starting with "virt-" and IP ending with ".84":**
    ```bash
    kubectl-sql -n default "select * from pods where name ~= '^virt-' and status.podIP ~= '[.]84$'"
    NAMESPACE	NAME                          	PHASE  	hostIP        	CREATION_TIME(RFC3339)       	
    default  	virt-launcher-test-bdw2p-lcrwx	Running	192.168.126.56	2020-02-12T14:14:01+02:00
    ...
    ```

---

**SI Units**

* **Get PVCs less than 20Gi and output as JSON:**
    ```bash
    kubectl-sql -o json "select * from */pvc where spec.resources.requests.storage < 20Gi"

    ...  json
    {
      "storage": "10Gi"
    }
    ...
    ```

---

**Comparing fields**

* **Get replica sets with 3 replicas but less ready replicas:**
    ```bash
    kubectl-sql --all-namespaces "select * from rs where spec.replicas = 3 and status.readyReplicas < spec.replicas"

    ...
    ```

---

**Join**

<p align="center">
   <a href="https://asciinema.org/a/AiBPT3SL7R9MgHCJV1tI0k6fU" target="_blank"><img src="https://asciinema.org/a/AiBPT3SL7R9MgHCJV1tI0k6fU.svg" /></a>
<p>
  
* **Display non-running pods by nodes for all namespaces:**
    ```bash
    kubectl-sql "select * from nodes join pods on \
        nodes.status.addresses[1].address = pods.status.hostIP and not pods.phase ~= 'Running'" -A
    ...
    ```

---

**Escaping Identifiers**

* **Use square brackets for identifiers with special characters:**
    ```bash
    ./kubectl-sql --all-namespaces "select * from pods where name ~= 'cni' and metadata.labels[openshift.io/component] = 'network'"
    ...
    ```

---

**Print help**

* **Display kubectl-sql help:**
    ```bash
    kubectl-sql --help
    ...
    ```
