# Configuring higher workload density

> [!IMPORTANT]
> Plesae avoid using the `NodeSwap` feature gate and its corresponding `SwapBehavior` configuration in kubelet.
> Wasp-agent and node swap feature-gate are mutually exclusive.

You can configure a higher VM workload and pod workload density in your cluster 
by over-committing memory resources (RAM).

While over-committed memory can lead to a higher workload density, at
the same time this will lead to some side-effects:

- Lower workload performance on a highly utilized system

Some workloads are more suited for higher workload density than
others, for example:

- Many similar workloads
- Underutilized workloads

## Configuring higher workload density with the wasp-agent

[wasp-agent]  is a component that enables an OpenShift cluster to assign 
SWAP resources to burstable pod and VM workloads. 

SWAP usage is supported on worker nodes only.

### Prerequisites

* `oc` is available
* Logged into cluster with `cluster-admin` role
* A defined memory over-commit ratio. By default: 150%
* A worker pool

### Procedure

> [!NOTE]
> The `wasp-agent` will deploy an OCI hook and periodically 
> verifies the swap limit to ensure it is set correctly in order to enable
> swap usage for containers on the node level.
> The low-level nature requires the `DaemonSet` to be privileged.

1. #### Create a `KubeletConfiguration` according to the following [example](../manifests/openshift/kubelet-configuration-with-swap.yaml).

2. #### Wait for the worker nodes to sync with the new config:
```console
$ oc wait mcp worker --for condition=Updated=True --timeout=-1s
```
3. #### Create `MachineConfig` to provision swap according to the following [example](../manifests/openshift/machineconfig-add-swap.yaml)

> [!IMPORTANT]
> In order to have enough swap for the worst case scenario, it must
> be ensured to have at least as much swap space provisioned as RAM
> is being over-committed.
> The amount of swap space to be provisioned on a node must
> be calculated according to the following formula:
>
>     NODE_SWAP_SPACE = NODE_RAM * (MEMORY_OVER_COMMIT_PERCENT / 100% - 1)
>
> Example:
>
>     NODE_SWAP_SPACE = 16 GB * (150% / 100% - 1)
>                     = 16 GB * (1.5 - 1)
>                     = 16 GB * (0.5)
>                     =  8 GB

4. #### Create a `MachineConfig` according to the following [example](../manifests/openshift/machineconfig-add-swap.yaml).

5. #### Create a privileged service account:

```console
$ oc adm new-project wasp
$ oc create sa -n wasp wasp
$ oc create clusterrolebinding wasp --clusterrole=cluster-admin --serviceaccount=wasp:wasp
$ oc adm policy add-scc-to-user -n wasp privileged -z wasp
```

6. #### Wait for the worker nodes to sync with the new config:
```console
$ oc wait mcp worker --for condition=Updated=True --timeout=-1s
```

7. #### Deploy `wasp-agent` <br>
 * ##### Determine wasp-agent image pull URL:
```console
oc get csv -n openshift-cnv -l=operators.coreos.com/kubevirt-hyperconverged.openshift-cnv -ojson | jq '.items[0].spec.relatedImages[] | select(.name|test(".*wasp-agent.*")) | .image'
```
  * ##### Create a `DaemonSet` with the relevant image URL according to the following [example](../manifests/openshift/ds.yaml).

8. #### Deploy alerting rules according to the following [example](../manifests/openshift/prometheus-rule.yaml) and add the cluster-monitoring label to the wasp namespace.
```console
$ oc label namespace wasp openshift.io/cluster-monitoring="true"
```

9. #### Configure OpenShift Virtualization to use memory overcommit using

   a. Via the OpenShift Console: <br>
       **Virtualization -> Overview -> Settings -> General Settings -> Memory Density** <br>
       ![image](https://github.com/user-attachments/assets/07c02c7c-0cd6-4377-9119-a8f3b7a58695)

   b. Alternatively via the CLI: [HCO example](../manifests/openshift/hco-set-memory-overcommit.yaml):

```console
$ oc patch --type=merge \
  -f <../manifests/openshift/hco-set-memory-overcommit.yaml> \
  --patch-file <../manifests/openshift/hco-set-memory-overcommit.yaml>
```

### Upgrade path
For users of wasp-agent v1.0, which lacks LimitedSwap, here is the upgrade path:
1. #### Adjust KubeletConfig:
   If you have installed the KubeletConfiguration object from version v1.0, you need to update it. 
   This can be done by applying the updated [KubeletConfiguration example](../manifests/openshift/kubelet-configuration-with-swap.yaml).
```console
$ oc replace -f <../manifests/openshift/kubelet-configuration-with-swap.yaml>
```
2. #### Update DaemonSet (make sure the image URL is valid):
   Apply the updated DaemonSet by using the provided [DaemonSet example](../manifests/openshift/ds.yaml).
```console
$ oc apply -f <../manifests/openshift/ds.yaml>
```
3. #### Update monitoring object:
```console
$ oc delete AlertingRule wasp-alerts -nopenshift-monitoring
$ oc create -f <../manifests/openshift/prometheus-rule.yaml>
```

### Verification

1. Validate the deployment
   TBD
2. Validate correctly provisioned swap by running:

       $ oc get nodes -l node-role.kubernetes.io/worker
       # Select a node from the provided list

       $ oc debug node/<selected-node> -- free -m

   Should show an amoutn larger than zero for swap, similar to:

                      total        used        free      shared  buff/cache   available
       Mem:           31846       23155        1044        6014       14483        8690
       Swap:           8191        2337        5854


3. Validate OpenShift Virtualization memory overcommitment configuration
   by running:

       $ oc get -n openshift-cnv HyperConverged kubevirt-hyperconverged -o jsonpath="{.spec.higherWorkloadDensity.memoryOvercommitPercentage}"
       150

   The returned value (in this case `150`) should match the value you
   have configured earlier on.

4. Validate Virtual Machine memory overcommitment
   TBD

### Additional Resources

[wasp-agent]: https://github.com/openshift-virtualization/wasp-agent
FPR: Free-Page Reporting
KSM:
