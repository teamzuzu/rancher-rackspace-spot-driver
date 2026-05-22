# Installation

This guide walks through installing the Rackspace Spot driver in Rancher and creating your first cluster.

---

## Prerequisites

- A running **Rancher** instance (v2.10 or later — UI extension requires Dashboard v2.10+)
- A **Rackspace Spot** account with an organization already created
- A Rackspace Spot **refresh token** (see [Obtaining credentials](#obtaining-credentials) below)
- Outbound HTTPS access from your Rancher server to `api.rackspace.com`

---

## Obtaining credentials

The driver authenticates with Rackspace Spot using a long-lived **refresh token**.

1. Log in to the [Rackspace Spot console](https://spot.rackspace.com)
2. Navigate to **Account → API Access**
3. Click **Get new token**, give it a name, and copy the value — you will not see it again

!!! warning "Keep your token secret"
    Store the token in Rancher as a secret (the UI marks the field as a password). Do not commit it to source control.

---

## Installing the driver

The driver and its UI extension are distributed as a Helm chart via the GitHub Pages Helm repository.

### Step 1 — Add the extension repository

1. Open the Rancher UI and navigate to **☰ → Extensions**
2. Click **⋮ → Manage Extension Repositories**
3. Click **Create** and fill in:

   | Field | Value |
   |---|---|
   | **Name** | `rackspace-spot` |
   | **URL** | `https://teamzuzu.github.io/rancher-rackspace-spot-driver` |

4. Click **Create**

Alternatively, install via Helm directly:

```bash
helm repo add rackspace-spot https://teamzuzu.github.io/rancher-rackspace-spot-driver
helm repo update
helm install rackspacespot rackspace-spot/rackspacespot \
  --namespace cattle-ui-plugin-system \
  --create-namespace
```

### Step 2 — Install the extension

1. Navigate to **☰ → Extensions → Available**
2. Find **Rackspace Spot** and click **Install**

Rancher installs the Helm chart, which registers:

- A `KontainerDriver` — tells Rancher where to download the binary
- A `UIPlugin` — tells Rancher's Dashboard where to load the cluster configuration form from

The cluster driver status changes to **Active** within ~30 seconds.

---

## Creating a cluster

1. Go to **☰ → Cluster Management → Clusters → Create**
2. Select **Rackspace Spot** from the provider list
3. Fill in the cluster configuration (see [Configuration reference](configuration.md)):

### Required fields

| Field | Description |
|---|---|
| **Cluster Name** | A display name for the cluster in Rancher (automatically converted to a valid CloudSpace name) |
| **Organization** | Your Rackspace Spot organization name |
| **Refresh Token** | Your Rackspace Spot API refresh token |

### Optional fields

See the full [Configuration reference](configuration.md) for all available options.

4. Click **Create** and wait for the cluster to reach the **Active** state

!!! info "Provisioning time"
    A new CloudSpace typically takes 5–25 minutes to reach the `Active` state ⏳ — grab a coffee while Rancher polls and updates the cluster status automatically.

---

## Upgrading the driver

To upgrade, update the repository and upgrade the Helm chart:

```bash
helm repo update rackspace-spot
helm upgrade rackspacespot rackspace-spot/rackspacespot \
  --namespace cattle-ui-plugin-system
```

Or via **☰ → Extensions → Installed → Rackspace Spot → ⋮ → Upgrade**.

Existing clusters are not affected until their next reconciliation.

---

## Removing the driver

1. Navigate to **☰ → Extensions → Installed**
2. Find **Rackspace Spot** and click **⋮ → Uninstall**

!!! warning
    Delete all Rackspace Spot clusters **before** removing the driver. If you remove the driver first, Rancher loses the ability to call the cloud API to clean up the underlying resources.

