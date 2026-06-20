# Import Existing Cluster Implementation Design

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Allow users to import an existing Rackspace Spot cloudspace into Rancher as a fully-managed cluster, giving them control over node pools via the Edit form after import.

**Approach:** An `importExistingCluster` boolean flag gates a separate code path in `Create`. When set, the driver skips pool validation and reconciliation, reads the real cloudspace config from the Spot API, maps it to `clusterState`, and returns. `PostCheck` runs unchanged — the cloudspace is already Active so `waitForCloudspace` returns immediately.

---

## Architecture

Two code paths in `Create`, selected by the `importExistingCluster` boolean option:

- **Normal path** (existing behaviour, unchanged): `stateFromOptions` → validate → `ensureCloudspace` → `reconcileSpotNodePools` → `reconcileOnDemandNodePool` → save
- **Import path**: minimal validation (org/token/name) → `GetCloudspace` → `stateFromCloudspace` → save → return (no reconciliation)

After import, `PostCheck` calls `waitForCloudspace` as usual; the cloudspace is already in Ready state so it returns immediately. The cluster becomes Active in Rancher and the Edit form shows the real pool config.

---

## Driver changes

### `GetDriverCreateOptions` (`driver/driver.go`)

Add one new boolean flag so Rancher records the value in cluster options:

```go
{
    Name:  "importExistingCluster",
    Label: "Import Existing Cluster",
    Type:  "boolean",
},
```

### `Create` (`driver/driver.go`)

Branch at the top of the method, before `stateFromOptions`:

```go
if getBoolOption(opts, "import-existing-cluster", "importExistingCluster") {
    return d.importCluster(ctx, opts, clusterInfo)
}
// existing flow unchanged below
```

### New `importCluster` method (`driver/driver.go`)

```go
func (d *Driver) importCluster(ctx context.Context, opts *types.DriverOptions, clusterInfo *types.ClusterInfo) (*types.ClusterInfo, error) {
    org   := getStringOption(opts, flagOrganization, "rackspaceSpotOrganization")
    token := getStringOption(opts, flagRefreshToken, "rackspaceSpotRefreshToken")
    if org == "" {
        return nil, fmt.Errorf("%s is required", flagOrganization)
    }
    if token == "" {
        return nil, fmt.Errorf("%s is required", flagRefreshToken)
    }

    rawName := opts.StringOptions["name"]
    cloudspaceName, err := sanitizeResourceName(rawName)
    if err != nil {
        return nil, fmt.Errorf("invalid cluster name: %w", err)
    }
    if cloudspaceName == "" {
        return nil, fmt.Errorf("cluster name is required")
    }

    client, err := newSpotClient(ctx, token, org)
    if err != nil {
        return nil, err
    }

    cs, err := client.api.GetCloudspace(ctx, org, cloudspaceName)
    if err != nil {
        if isNotFound(err) {
            return nil, fmt.Errorf("cloudspace %q not found: use create mode to provision a new cluster", cloudspaceName)
        }
        return nil, fmt.Errorf("failed to look up cloudspace %q: %w", cloudspaceName, err)
    }

    s := stateFromCloudspace(cs, org, token)

    info := clusterInfo
    if info == nil {
        info = &types.ClusterInfo{}
    }
    if err := s.save(info); err != nil {
        return info, err
    }

    logrus.Infof("[%s] imported cloudspace %s (region: %s, k8s: %s, spot pools: %d)",
        driverName, cloudspaceName, s.Region, s.KubernetesVersion,
        1+len(s.AdditionalSpotPools))

    return info, nil
}
```

### New `stateFromCloudspace` (`driver/config.go`)

Maps a `*spotv1.CloudSpace` (plus org and token) to a `*clusterState`. All fields that exist on the cloudspace are mapped; org and token come from the user-supplied opts since they are not stored on the cloudspace itself.

```go
func stateFromCloudspace(cs *spotv1.CloudSpace, org, token string) *clusterState {
    s := &clusterState{
        RefreshToken:      token,
        Organization:      org,
        CloudspaceName:    cs.Name,
        Region:            cs.Region,
        KubernetesVersion: cs.KubernetesVersion,
        CNI:               cs.CNI,
        GPUEnabled:        cs.GpuEnabled,
        PreemptionWebhook: cs.PreemptionWebhookURL,
        DeploymentType:    cs.DeploymentType,
    }

    // Map spot node pools: first → primary, rest → AdditionalSpotPools
    for i, p := range cs.SpotNodepools {
        if i == 0 {
            s.SpotPoolName    = p.Name
            s.SpotServerClass = p.ServerClass
            s.SpotNodeCount   = p.Desired
            s.SpotBidPrice    = p.BidPrice
            s.SpotAutoscaling = p.Autoscaling.Enabled
            s.SpotMinNodes    = p.Autoscaling.MinNodes
            s.SpotMaxNodes    = p.Autoscaling.MaxNodes
        } else {
            s.AdditionalSpotPools = append(s.AdditionalSpotPools, SpotPoolConfig{
                Name:        p.Name,
                ServerClass: p.ServerClass,
                NodeCount:   p.Desired,
                BidPrice:    p.BidPrice,
                Autoscaling: p.Autoscaling.Enabled,
                MinNodes:    p.Autoscaling.MinNodes,
                MaxNodes:    p.Autoscaling.MaxNodes,
            })
        }
    }

    // Map first on-demand pool (at most one is supported)
    if len(cs.OnDemandNodePools) > 0 {
        p := cs.OnDemandNodePools[0]
        s.OnDemandEnabled  = true
        s.OnDemandPoolName = p.Name
        s.OnDemandClass    = p.ServerClass
        s.OnDemandCount    = p.Desired
        s.OnDemandPrice    = p.OnDemandPricePerHour
    }

    return s
}
```

**Edge cases:**
- Cloudspace with no spot pools → primary pool fields are zero-valued; user adds pools via Edit
- Cloudspace with no on-demand pool → `OnDemandEnabled=false` (zero value, correct)
- Cloudspace in non-Ready state → `PostCheck`'s `waitForCloudspace` surfaces the status and times out normally

---

## UI changes (`ui/pkg/rackspacespot/components/CruRackspaceSpot.vue`)

1. Add `importExistingCluster` boolean to the reactive `config` object (default `false`)
2. Add a toggle labelled **"Import existing cluster"** near the top of the form, below the credentials fields
3. When `importExistingCluster` is `true`: hide the primary spot pool section, autoscaling config, on-demand pool section, and additional pools section
4. Show a note beneath the toggle: *"Node pool configuration will be read from the existing cloudspace after import."*
5. The existing **Cluster Name** field doubles as the cloudspace lookup key — no new field needed

The `importExistingCluster` value is submitted as a boolean option; the driver reads it via `getBoolOption(opts, "importExistingCluster")`.

---

## Error handling

| Condition | Behaviour |
|---|---|
| Import mode, cloudspace not found | Error: `"cloudspace %q not found: use create mode to provision a new cluster"` |
| Import mode, org or token empty | Same required-field errors as normal create |
| Import mode, cluster name empty or invalid | Error from `sanitizeResourceName` |
| Import mode, cloudspace has no spot pools | Import succeeds; zero pool fields in state; user adds pools via Edit |
| Import mode, cloudspace in error/non-Ready state | `PostCheck`'s `waitForCloudspace` times out with current status message |

---

## Testing

### `driver/config_test.go`

- `TestStateFromCloudspace_fullConfig` — cloudspace with 2 spot pools + 1 on-demand pool; verify all fields map correctly, first pool → primary, second → `AdditionalSpotPools[0]`
- `TestStateFromCloudspace_noOnDemand` — no on-demand pools; `OnDemandEnabled=false`
- `TestStateFromCloudspace_singleSpotPool` — one spot pool; `AdditionalSpotPools` is nil/empty

### `driver/driver_test.go` (or new `driver/import_test.go`)

- `TestImportCluster_notFound` — `GetCloudspace` returns not-found error; assert returned error contains "not found"
- `TestImportCluster_success` — `GetCloudspace` returns a cloudspace; assert `ClusterInfo.Metadata` contains correctly serialized state; assert no reconcile methods were called on the mock client

### `ui/pkg/rackspacespot/test.mjs`

- When `importExistingCluster` is `true`, the pool-related config fields are not required / form sections are hidden — pure logic test, no browser needed
