# Configuration Reference

All options are available in the Rancher UI when creating or editing a cluster. This page documents every field, its type, default value, and when to use it.

---

## Authentication

| Option | Type | Required | Description |
|---|---|---|---|
| `refreshToken` | string (password) | Yes | Rackspace Spot API refresh token. Stored as a Rancher secret and never logged. |
| `organization` | string | Yes | Rackspace Spot organization name. Must match exactly as shown in the Spot console. |

---

## Cluster settings

| Option | Type | Default | Description |
|---|---|---|---|
| `region` | string | **required** | Rackspace Spot region where the CloudSpace is created (e.g. `us-east-1`). |
| `k8sVersion` | string | `1.33.0` | Kubernetes version for the control plane. Must be a version supported by Rackspace Spot. |
| `cni` | string | `calico` | CNI plugin. Accepted values: `calico`, `cilium`, `byocni`. |
| `gpuEnabled` | bool | `false` | Attach GPU resources to the CloudSpace. Requires a region that supports GPUs. |
| `preemptionWebhook` | string | — | HTTPS URL called by Rackspace Spot before preempting a spot node. Useful for draining workloads gracefully. |
| `deploymentType` | string | — | Override the CloudSpace deployment type (e.g. `spot`, `on-demand`). Leave blank to use the region default. |

---

## Spot node pool

The spot node pool is always created. It is the primary source of compute for the cluster.

| Option | Type | Default | Description |
|---|---|---|---|
| `spotPoolName` | string | *(auto-generated UUID)* | Name for the spot node pool. Must be unique within the CloudSpace. Auto-generated if not provided. |
| `spotServerClass` | string | `gh.vs1.medium-iad` | [Server class](https://spot.rackspace.com) for spot nodes (CPU/RAM/GPU tier). |
| `spotNodeCount` | int | `3` | Desired number of spot nodes. Ignored when autoscaling is enabled. |
| `spotBidPrice` | string | `0.01` | Maximum price per node-hour in USD. Nodes are only provisioned when the market price is at or below this value. |
| `spotAutoscaling` | bool | `false` | Enable autoscaling for the spot pool. When enabled, `spotNodeCount` is overridden by the autoscaler. |
| `spotMinNodes` | int | `1` | Minimum node count when autoscaling is enabled. |
| `spotMaxNodes` | int | `10` | Maximum node count when autoscaling is enabled. |

### Bid price strategy

!!! tip "Choosing a bid price"
    - Set your bid price to the **on-demand equivalent** of your workload to maximise availability.
    - Set a **lower bid price** to cap costs — nodes are not provisioned if the market price exceeds your bid.
    - Rackspace Spot publishes current market prices in the console under **Spot Market**.

---

## On-demand node pool

An optional second pool that uses reserved capacity. Use it for workloads that cannot tolerate preemption (e.g. stateful services, CI orchestrators).

| Option | Type | Default | Description |
|---|---|---|---|
| `onDemandEnabled` | bool | `false` | When `true`, the on-demand pool is created (or updated). When `false`, any existing on-demand pools are deleted. |
| `onDemandPoolName` | string | *(auto-generated UUID)* | Name for the on-demand node pool. Auto-generated if not provided. |
| `onDemandClass` | string | — | Server class for on-demand nodes. Required when `onDemandEnabled` is `true`. |
| `onDemandCount` | int | `1` | Desired number of on-demand nodes. |
| `onDemandPrice` | string | — | Maximum price per node-hour in USD. Optional — leave blank to use the region list price. |

---

## Additional spot pools

You can define additional spot node pools beyond the primary one by providing a JSON array in `additionalSpotPools`. Each entry supports the same fields as the primary spot pool:

```json
[
  {
    "name":        "",
    "serverClass": "gh.vs1.medium-iad",
    "nodeCount":   2,
    "bidPrice":    "0.01",
    "autoscaling": false,
    "minNodes":    1,
    "maxNodes":    10
  }
]
```

If `name` is blank or not a valid UUID it is auto-generated. Pools that are present in the running cluster but absent from the desired list are deleted on the next update.

---

## Update options

When editing an existing cluster, only the following fields are reconciled. All other fields require cluster recreation.

| Option | Updateable |
|---|---|
| `refreshToken` | Yes (use to rotate credentials without recreating the cluster) |
| `k8sVersion` | Yes (stored; applied on next reconciliation) |
| `spotNodeCount` | Yes |
| `spotBidPrice` | Yes |
| `spotAutoscaling` | Yes |
| `spotMinNodes` | Yes |
| `spotMaxNodes` | Yes |
| `onDemandEnabled` | Yes (enables or disables the on-demand pool; disabling deletes existing pools) |
| `onDemandCount` | Yes |
| `additionalSpotPools` | Yes (adds, updates, or removes additional spot pools) |

---

## Cluster name sanitization

Rancher allows cluster names that contain spaces, uppercase letters, and special characters. The driver automatically converts the Rancher cluster name to a valid Rackspace Spot CloudSpace name by:

1. Converting to lowercase
2. Replacing any character that is not `a-z`, `0-9`, or `-` with a hyphen
3. Trimming leading and trailing hyphens
4. Truncating to 63 characters

For example, `My Cluster (prod)` becomes `my-cluster--prod-`.

!!! warning
    The sanitized name is stored in the cluster state after creation. Renaming a cluster in Rancher does not rename the underlying CloudSpace.

---

## Example: autoscaling spot + on-demand pool

```
refreshToken:        <your token>
organization:        acme-corp
region:              <your-region>
k8sVersion:          1.33.0
cni:                 calico

spotServerClass:     gh.vs1.medium-iad
spotBidPrice:        0.01
spotAutoscaling:     true
spotMinNodes:        2
spotMaxNodes:        20

onDemandEnabled:     true
onDemandClass:       rxtx.2xlarge
onDemandCount:       2
```

This configuration creates a cluster with:

- A spot pool that autoscales between 2 and 20 `gh.vs1.medium-iad` nodes, capped at $0.01/hr per node
- A fixed on-demand pool of 2 `rxtx.2xlarge` nodes for guaranteed capacity
- Pool names are auto-generated UUIDs
