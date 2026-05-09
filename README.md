# Rancher Cluster Driver — Rackspace Spot

A [Rancher kontainer-engine](https://github.com/rancher/kontainer-engine) cluster driver that provisions and manages Kubernetes clusters on [Rackspace Spot](https://spot.rackspace.com) — the GPU-accelerated spot compute platform.

## Overview

This driver lets you create, scale, and delete Rackspace Spot **CloudSpaces** (managed Kubernetes clusters) directly from the Rancher UI or API, alongside your other cluster providers. It supports:

- Spot node pools with configurable bid prices and autoscaling
- Optional on-demand node pools for stable baseline capacity
- GPU-enabled CloudSpaces
- Preemption webhook integration
- Kubernetes version management

---

## Prerequisites

| Requirement | Details |
|---|---|
| Rancher | v2.6 or later |
| Rackspace Spot account | With at least one organization created |
| Rackspace Spot refresh token | See [Getting a refresh token](#getting-a-refresh-token) |

### Getting a refresh token

Install the Rackspace Spot CLI and authenticate:

```bash
# Install (see https://spot.rackspace.com/docs for latest instructions)
curl -sSL https://spot.rackspace.com/install | sh

# Log in — this stores a refresh token locally
rxtspot login

# Print your refresh token
rxtspot token
```

Keep the refresh token — you'll paste it into Rancher when creating a cluster.

---

## Installing the driver in Rancher

### Step 1 — Add the cluster driver

1. Open the Rancher UI and navigate to **☰ → Cluster Management**.
2. Select **Drivers** from the left sidebar, then click the **Cluster Drivers** tab.
3. Click **Add Cluster Driver**.
4. Fill in the form:

   | Field | Value |
   |---|---|
   | **Download URL** | `https://github.com/teamzuzu/rancher-rackspace-spot-driver/releases/download/v0.1.0/rancher-rackspace-spot-driver_linux_amd64.tar.gz` |
   | **Custom UI URL** | *(leave blank)* |
   | **Whitelist Domains** | `spot.rackspace.com` |

   > Replace `v0.1.0` with the [latest release](https://github.com/teamzuzu/rancher-rackspace-spot-driver/releases/latest) tag.

5. Click **Create** and wait for the driver status to change to **Active** (takes ~30 seconds while Rancher downloads and validates the binary).

### Step 2 — Create a cluster

1. Navigate to **☰ → Cluster Management → Clusters → Create**.
2. Choose **Rackspace Spot** from the provider list.
3. Fill in the cluster configuration (see [Configuration reference](#configuration-reference) below).
4. Click **Create**.

Rancher will call the driver, which creates the CloudSpace and node pools. The cluster moves through `Provisioning → Waiting → Active` as the control plane and nodes come online (typically 5–10 minutes).

### Upgrading the driver

To upgrade to a newer driver version, edit the cluster driver (Rancher UI → Cluster Drivers → ⋮ → Edit) and update the **Download URL** to the new release tarball. Existing clusters are unaffected — the driver version only matters when Rancher makes a new API call.

---

## Configuration reference

All options are set in the Rancher cluster creation form.

### Authentication

| Option | Required | Default | Description |
|---|---|---|---|
| `rackspace-spot-refresh-token` | ✅ | — | Rackspace Spot refresh token (stored as a secret) |
| `rackspace-spot-organization` | ✅ | — | Your Rackspace Spot organization name |

### Cluster

| Option | Required | Default | Description |
|---|---|---|---|
| `rackspace-spot-region` | | `colo-lax-1` | Rackspace Spot region |
| `kubernetes-version` | | `1.28` | Kubernetes version for the CloudSpace |
| `cni` | | `calico` | CNI plugin (`calico` or `flannel`) |
| `gpu-enabled` | | `false` | Enable GPU support on the CloudSpace |
| `preemption-webhook-url` | | — | Webhook called before spot nodes are preempted |
| `deployment-type` | | — | CloudSpace deployment type (provider-specific) |

### Spot node pool

A spot pool is always created. Spot nodes are preemptible and billed at your bid price when capacity is available.

| Option | Required | Default | Description |
|---|---|---|---|
| `spot-node-pool-name` | | `spot-pool` | Name for the spot node pool |
| `spot-server-class` | | `rxtx.4xlarge-mi300x` | Hardware class for spot nodes |
| `spot-node-count` | | `3` | Desired number of spot nodes |
| `spot-bid-price` | | `0.50` | Maximum price per node-hour (USD) |
| `spot-autoscaling-enabled` | | `false` | Enable autoscaling for the spot pool |
| `spot-autoscaling-min-nodes` | | `1` | Autoscaling lower bound |
| `spot-autoscaling-max-nodes` | | `10` | Autoscaling upper bound |

### On-demand node pool (optional)

An on-demand pool provides stable capacity that is never preempted. Useful for system workloads and daemonsets.

| Option | Required | Default | Description |
|---|---|---|---|
| `on-demand-enabled` | | `false` | Create an on-demand node pool |
| `on-demand-node-pool-name` | | `on-demand-pool` | Name for the on-demand pool |
| `on-demand-server-class` | | — | Hardware class for on-demand nodes |
| `on-demand-node-count` | | `1` | Desired number of on-demand nodes |
| `on-demand-price-per-hour` | | — | Maximum price per node-hour (USD) |

### Available server classes

Run `rxtspot server-classes list` to see classes available in your region, or visit the [Rackspace Spot portal](https://spot.rackspace.com).

---

## Building from source

```bash
git clone https://github.com/teamzuzu/rancher-rackspace-spot-driver
cd rancher-rackspace-spot-driver

# Resolve dependencies (requires Go 1.21+)
go mod tidy

# Build
make build
# Binary written to bin/rancher-rackspace-spot-driver

# Run tests
make test

# Build container image
make image
```

### Local driver testing with Rancher

You can host the driver binary on any HTTPS server and point Rancher at it. A quick way during development:

```bash
# Build and serve locally (requires ngrok or similar for HTTPS)
make build
python3 -m http.server 9999 --directory bin/
ngrok http 9999
# Use the ngrok HTTPS URL as the download URL in Rancher
```

---

## Architecture

```
Rancher → gRPC → rancher-rackspace-spot-driver
                        │
                        ├─ Create  → CreateCloudspace + CreateSpotNodePool
                        ├─ PostCheck → WaitForRunning + GetKubeconfig + ServiceAccount
                        ├─ Update  → UpdateSpotNodePool (scale / bid price)
                        └─ Remove  → DeleteNodePools + DeleteCloudspace
```

The driver binary is started by Rancher as a sidecar process. Rancher communicates with it over a local gRPC socket, then terminates it when the operation completes.

State (organization, cloudspace name, pool configuration) is serialised as JSON into `ClusterInfo.Metadata` and passed back on every subsequent call, so the driver is fully stateless between invocations.

---

## Releases

Binaries for `linux/amd64` and `linux/arm64` are published to [GitHub Releases](https://github.com/teamzuzu/rancher-rackspace-spot-driver/releases) automatically on every version tag. A container image is also pushed to the GitHub Container Registry:

```bash
docker pull ghcr.io/teamzuzu/rancher-rackspace-spot-driver:latest
```

---

## License

Apache 2.0 — see [LICENSE](LICENSE).
