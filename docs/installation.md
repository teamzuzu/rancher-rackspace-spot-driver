# Installation

This guide walks through registering the Rackspace Spot driver in Rancher and creating your first cluster.

---

## Prerequisites

- A running **Rancher** instance (v2.6 or later)
- A **Rackspace Spot** account with an organization already created
- A Rackspace Spot **refresh token** (see [Obtaining credentials](#obtaining-credentials) below)
- Outbound HTTPS access from your Rancher server to `api.rackspacespot.com`

---

## Obtaining credentials

The driver authenticates with Rackspace Spot using a long-lived **refresh token**.

1. Log in to the [Rackspace Spot console](https://console.rackspacespot.com)
2. Navigate to **Account → API tokens**
3. Click **Generate token** and copy the value — you will not see it again

!!! warning "Keep your token secret"
    Store the token in Rancher as a secret (the UI marks the field as a password). Do not commit it to source control.

---

## Registering the driver in Rancher

### 1. Add the cluster driver

1. In Rancher, open the **☰** menu and go to **Cluster Management**
2. Click **Drivers** in the left sidebar, then select the **Cluster Drivers** tab
3. Click **Add Cluster Driver** (top right)

Fill in the form:

| Field | Value |
|---|---|
| **Download URL** | `https://github.com/teamzuzu/rancher-rackspace-spot-driver/releases/latest/download/rancher-rackspace-spot-driver_linux_amd64.tar.gz` |
| **Custom UI URL** | `https://teamzuzu.github.io/rancher-rackspace-spot-driver/ui/component.js` |
| **Checksum** | *(optional — copy from the release checksums file)* |
| **Whitelist Domains** | `rackspacespot.com` |

!!! info "Custom UI"
    The Custom UI URL loads the cluster creation form directly in Rancher's Dashboard. Without it, Rancher's newer Dashboard UI does not show configuration fields. The legacy Cluster Manager UI works without a Custom UI URL.

Click **Create**.

### 2. Wait for activation

Rancher downloads and verifies the driver binary. The status column shows **Active** once registration is complete (usually under 30 seconds).

!!! note "arm64 nodes"
    If your Rancher server runs on arm64, change `amd64` to `arm64` in the download URL.

---

## Creating a cluster

1. Go to **Cluster Management → Create**
2. Select **Rackspace Spot** from the provider list
3. Fill in the cluster details:

### Required fields

| Field | Description |
|---|---|
| **Cluster Name** | A display name for the cluster in Rancher (converted to a valid CloudSpace name automatically) |
| **Refresh Token** | Your Rackspace Spot API refresh token |
| **Organization** | Your Rackspace Spot organization name |

### Optional fields

See the full [Configuration reference](configuration.md) for all available options.

4. Click **Create** and wait for the cluster to reach the **Active** state

!!! info "Provisioning time"
    A new CloudSpace typically takes 5–10 minutes to reach the `Running` state. Rancher polls every 30 seconds and updates the cluster status automatically.

---

## Upgrading the driver

To use a newer driver version, update the **Download URL** on the cluster driver's edit page to point to the new release asset, then click **Save**. Rancher will re-download and reload the binary.

Existing clusters are not affected by the driver upgrade until their next reconciliation.

---

## Removing the driver

1. Go to **Cluster Management → Drivers → Cluster Drivers**
2. Find the Rackspace Spot driver, click **⋮ → Delete**

!!! warning
    Delete all Rackspace Spot clusters **before** removing the driver. If you remove the driver first, Rancher loses the ability to call the cloud API to clean up the underlying resources.

---

## Installing via container (advanced)

If you manage Rancher with Helm, you can inject the driver binary as an init container and mount it into the Rancher pod. This is useful in air-gapped environments where Rancher cannot reach GitHub.

```yaml
# values.yaml excerpt
extraInitContainers:
  - name: spot-driver
    image: ghcr.io/teamzuzu/rancher-rackspace-spot-driver:latest
    command: ["/bin/sh", "-c", "cp /rancher-rackspace-spot-driver /drivers/"]
    volumeMounts:
      - name: drivers
        mountPath: /drivers

extraVolumeMounts:
  - name: drivers
    mountPath: /var/lib/rancher/kontainer-engine/drivers
```

Set the driver's **Download URL** to `file:///var/lib/rancher/kontainer-engine/drivers/rancher-rackspace-spot-driver` in Rancher after mounting.
