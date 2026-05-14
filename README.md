# Rancher Cluster Driver For Rackspace Spot



A [rancher kontainer-engine](https://github.com/rancher/kontainer-engine) cluster driver that provisions and manages Kubernetes clusters on [Rackspace Spot](https://spot.rackspace.com) — the spot compute platform.

## Overview

This driver lets you create, scale, and delete Rackspace Spot **CloudSpaces** (managed Kubernetes clusters) directly from the Rancher UI or API, alongside your other cluster providers. It supports:

- Spot node pools with configurable bid prices and autoscaling
- Optional on-demand node pools for stable baseline capacity
- GPU-enabled CloudSpaces
- Preemption webhook integration
- Kubernetes version management


[![CI](https://github.com/teamzuzu/rancher-rackspace-spot-driver/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/teamzuzu/rancher-rackspace-spot-driver/actions/workflows/ci.yml) [![Release](https://github.com/teamzuzu/rancher-rackspace-spot-driver/actions/workflows/release.yml/badge.svg?branch=main)](https://github.com/teamzuzu/rancher-rackspace-spot-driver/actions/workflows/release.yml) [![Deploy Docs](https://github.com/teamzuzu/rancher-rackspace-spot-driver/actions/workflows/pages.yml/badge.svg?branch=main)](https://github.com/teamzuzu/rancher-rackspace-spot-driver/actions/workflows/pages.yml)
---

## Prerequisites

| Requirement | Details |
|---|---|
| Rancher | v2.10 or later (UI extension requires Dashboard v2.10+) |
| Rackspace Spot account | With at least one organization created |
| Rackspace Spot refresh token | See [Getting a refresh token](#getting-a-refresh-token) |

### Getting a refresh token

The easiest way is directly from the Rackspace Spot console — no CLI required:

1. Log in to the [Rackspace Spot console](https://spot.rackspace.com)
2. Go to **Account → API Access**
3. Click **Get new token**, give it a name, and copy the value

Alternatively, use the [`spotctl`](https://github.com/rackspace-spot/spotctl) CLI:

```bash
# Download the latest release for your platform from:
# https://github.com/rackspace-spot/spotctl/releases

# Run the interactive setup — it will prompt for your token and org
spotctl configure
```

Keep the refresh token — you'll paste it into Rancher when creating a cluster.

---

## Installing the driver in Rancher

The driver and its UI extension are distributed as a Helm chart via a GitHub Pages Helm repository.

### Step 1 — Add the extension repository

1. Open the Rancher UI and navigate to **☰ → Extensions**.
2. Click **⋮ → Manage Extension Repositories**.
3. Click **Create** and fill in:

   | Field | Value |
   |---|---|
   | **Name** | `rackspace-spot` |
   | **URL** | `https://teamzuzu.github.io/rancher-rackspace-spot-driver` |

4. Click **Create**.

Alternatively, install via Helm directly:

```bash
helm repo add rackspace-spot https://teamzuzu.github.io/rancher-rackspace-spot-driver
helm repo update
helm install rackspacespot rackspace-spot/rackspacespot \
  --namespace cattle-ui-plugin-system \
  --create-namespace
```

### Step 2 — Install the extension

1. Navigate to **☰ → Extensions → Available**.
2. Find **Rackspace Spot** and click **Install**.

Rancher installs the Helm chart, which registers:
- A `KontainerDriver` — tells Rancher where to download the binary
- A `UIPlugin` — tells Rancher's Dashboard where to load the UI extension from

The cluster driver status changes to **Active** within ~30 seconds.

### Step 3 — Create a cluster

1. Navigate to **☰ → Cluster Management → Clusters → Create**.
2. Choose **Rackspace Spot** from the provider list.
3. Fill in the cluster configuration (see [Configuration reference](#configuration-reference) below).
4. Click **Create**.

Rancher will call the driver, which creates the CloudSpace and node pools. The cluster moves through `Provisioning → Waiting → Active` as the control plane and nodes come online (typically 5–10 minutes).

### Upgrading the driver

To upgrade, update the repository and upgrade the Helm chart:

```bash
helm repo update rackspace-spot
helm upgrade rackspacespot rackspace-spot/rackspacespot \
  --namespace cattle-ui-plugin-system
```

Or via **☰ → Extensions → Installed → Rackspace Spot → ⋮ → Upgrade**.

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
| `kubernetes-version` | | `1.32.9` | Kubernetes version for the CloudSpace |
| `cni` | | `calico` | CNI plugin (`calico`, `cilium`, or `byocni`) |
| `gpu-enabled` | | `false` | Enable GPU support on the CloudSpace |
| `preemption-webhook-url` | | — | Webhook called before spot nodes are preempted |
| `deployment-type` | | — | CloudSpace deployment type (provider-specific) |

### Spot node pool

A spot pool is always created. Spot nodes are preemptible and billed at your bid price when capacity is available.

| Option | Required | Default | Description |
|---|---|---|---|
| `spot-node-pool-name` | | *(auto-generated UUID)* | Name for the spot node pool |
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
| `on-demand-node-pool-name` | | *(auto-generated UUID)* | Name for the on-demand pool |
| `on-demand-server-class` | | — | Hardware class for on-demand nodes |
| `on-demand-node-count` | | `1` | Desired number of on-demand nodes |
| `on-demand-price-per-hour` | | — | Maximum price per node-hour (USD) |

### Available server classes

Run `rxtspot server-classes list` to see classes available in your region, or visit the [Rackspace Spot portal](https://spot.rackspace.com).

---

## Building from source

### Driver binary

```bash
git clone https://github.com/teamzuzu/rancher-rackspace-spot-driver
cd rancher-rackspace-spot-driver

# Resolve dependencies (requires Go 1.24+)
go mod tidy

# Build
make build
# Binary written to bin/kontainer-engine-driver-rackspacespot

# Run tests
make test

# Build container image
make image
```

#### Local driver testing with Rancher

You can host the driver binary on any HTTPS server and point Rancher at it. A quick way during development:

```bash
# Build and serve locally (requires ngrok or similar for HTTPS)
make build
python3 -m http.server 9999 --directory bin/
ngrok http 9999
# Use the ngrok HTTPS URL as the download URL in Rancher
```

### UI extension

The cluster configuration form is a [Rancher UI Extension](https://extensions.rancher.io) built with Vue and [`@rancher/shell`](https://github.com/rancher/shell) 3.0.8. The source lives in `ui/pkg/rackspacespot/`.

```bash
cd ui
npm install --legacy-peer-deps

# Build the extension bundle (outputs to dist-pkg/rackspacespot-<version>/)
npm run build-pkg rackspacespot
```

The extension is distributed via the Helm repository and loaded by Rancher's UIPlugin mechanism — no manual file copying required. The release workflow (`release.yml`) builds the extension, packages the Helm chart, and publishes the extension files to GitHub Pages automatically on every release.

---

## Architecture

```
Browser (Rancher Dashboard)
  └─ Vue UI Extension (ui/pkg/rackspacespot/) ─── served from GitHub Pages
        │ writes genericEngineConfig
        ▼
Rancher API → gRPC → kontainer-engine-driver-rackspacespot (Go binary)
                              │
                              ├─ Create    → CreateCloudspace + CreateSpotNodePool
                              ├─ PostCheck → WaitForReady + GetKubeconfig + ServiceAccount
                              ├─ Update    → UpdateSpotNodePool (scale / bid price)
                              └─ Remove    → DeleteNodePools + DeleteCloudspace
```

The driver binary is started by Rancher as a sidecar process. Rancher communicates with it over a local gRPC socket, then terminates it when the operation completes.

State (organization, cloudspace name, pool configuration) is serialised as JSON into `ClusterInfo.Metadata` and passed back on every subsequent call, so the driver is fully stateless between invocations. The refresh token is stored separately in `ClusterInfo.Password` (not in the JSON blob) so it is never logged.

The UI extension (`ui/`) is a separate build artifact compiled with `@rancher/shell` 3.0.8. It is registered as a `kontainer` provisioner and writes cluster config into `cluster.genericEngineConfig`, which Rancher passes through to the driver binary.

---

## Releases

Binaries for `linux/amd64` and `linux/arm64` are published to [GitHub Releases](https://github.com/teamzuzu/rancher-rackspace-spot-driver/releases) automatically on every version tag. A container image is also pushed to the GitHub Container Registry:

```bash
docker pull ghcr.io/teamzuzu/rancher-rackspace-spot-driver:latest
```

---

## License

Apache 2.0 — see [LICENSE](LICENSE).
