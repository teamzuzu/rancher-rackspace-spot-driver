# Rancher Rackspace Spot Driver

**Deploy Kubernetes clusters on Rackspace Spot instances — directly from the Rancher UI.**

The Rancher Rackspace Spot Driver is a [kontainer-engine](https://github.com/rancher/kontainer-engine) plugin that lets you provision, manage, and delete Rackspace Spot CloudSpaces without leaving Rancher. Spot instances give you high-performance compute at significantly reduced cost, with optional on-demand node pools for workloads that need guaranteed capacity.

---

## Features

<div class="grid cards" markdown>

-   :material-lightning-bolt:{ .lg .middle } **One-click provisioning**

    ---

    Create a fully managed Kubernetes cluster on Rackspace Spot from the Rancher UI in minutes.

-   :material-currency-usd:{ .lg .middle } **Cost-efficient spot nodes**

    ---

    Set a bid price and let Rackspace Spot fill your node pool with spare capacity — pay a fraction of on-demand rates.

-   :material-scale-balance:{ .lg .middle } **Autoscaling built-in**

    ---

    Configure min/max node counts per pool and let the cluster scale automatically with your workload.

-   :material-server-plus:{ .lg .middle } **Hybrid node pools**

    ---

    Pair a spot pool with an optional on-demand pool for workloads that require guaranteed availability.

-   :material-kubernetes:{ .lg .middle } **Kubernetes version management**

    ---

    Select your Kubernetes version at creation time and upgrade in-place through the standard Rancher UI.

-   :material-shield-check:{ .lg .middle } **Secure by default**

    ---

    Credentials are stored as Rancher secrets. The driver creates a scoped `rancher` service account with the minimum required permissions.

</div>

---

## Quick start

=== "Install the driver"

    1. In Rancher, go to **☰ → Extensions**
    2. Click **⋮ → Manage Extension Repositories → Create**
    3. Set **Name** to `rackspace-spot` and **URL** to `https://teamzuzu.github.io/rancher-rackspace-spot-driver`
    4. Click **Create**, then go to **☰ → Extensions → Available**
    5. Find **Rackspace Spot** and click **Install**

    Or via Helm:

    ```bash
    helm repo add rackspace-spot https://teamzuzu.github.io/rancher-rackspace-spot-driver
    helm repo update
    helm install rackspacespot rackspace-spot/rackspacespot \
      --namespace cattle-ui-plugin-system --create-namespace
    ```

=== "Create a cluster"

    1. Go to **Cluster Management → Create**
    2. Select **Rackspace Spot** from the driver list
    3. Enter your **Refresh Token** and **Organization**
    4. Configure node pools and click **Create**

!!! tip "Full details"
    See the [Installation guide](installation.md) for step-by-step instructions, and the [Configuration reference](configuration.md) for all available options.

---

## Architecture

```
Rancher UI
    │
    ▼
rancher-rackspace-spot-driver  (gRPC plugin)
    │
    ├── Rackspace Spot API
    │       ├── CloudSpace (managed control plane)
    │       ├── SpotNodePool
    │       └── OnDemandNodePool
    │
    └── Cluster kubeconfig → Rancher service account bootstrap
```

The driver runs as a sidecar process inside the Rancher container. Rancher calls it over a local gRPC socket whenever you create, update, or delete a cluster.

---

## Requirements

| Requirement | Version |
|---|---|
| Rancher | ≥ 2.10 |
| Rackspace Spot account | Any |
| Go (to build from source) | ≥ 1.24 |

---

## License

Apache 2.0 — see [LICENSE](https://github.com/teamzuzu/rancher-rackspace-spot-driver/blob/main/LICENSE) for details.
