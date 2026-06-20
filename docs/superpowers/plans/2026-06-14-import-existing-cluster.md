# Import Existing Cluster Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow users to import an existing Rackspace Spot cloudspace into Rancher as a fully-managed cluster by setting an `importExistingCluster` flag, which routes `Create` through a separate code path that reads the real cloudspace config from the Spot API instead of provisioning new infrastructure.

**Architecture:** An `importExistingCluster` boolean flag in the driver options selects between the normal create path and a new import path. In import mode, `Create` calls a new `importCloudspace` method on `spotClient` that calls `GetCloudspace`, maps the response to `clusterState` via `stateFromCloudspace`, and returns — no pool reconciliation. `PostCheck` is unchanged. The Vue UI adds a toggle that hides pool config sections when import mode is active.

**Tech Stack:** Go 1.21, `spot-go-sdk v0.2.0`, Vue 3 / `@rancher/shell` 3.0.8, Node.js (UI tests).

---

## File Map

| File | Change |
|---|---|
| `driver/config.go` | Add `stateFromCloudspace` function |
| `driver/config_test.go` | Add 3 tests for `stateFromCloudspace` |
| `driver/client.go` | Add `importCloudspace` method on `spotClient` |
| `driver/import_test.go` | New file — `importMockAPI` + 2 tests for `importCloudspace` |
| `driver/driver.go` | Add `flagImportExisting` const; add flag to `GetDriverCreateOptions`; add import branch to `Create`; add `importCluster` method |
| `ui/pkg/rackspacespot/components/CruRackspaceSpot.vue` | Add import toggle; wrap pool sections with `v-if`; skip pool validation in `save()` |
| `ui/pkg/rackspacespot/test.mjs` | Add `validateConfig` function + 4 tests |

---

## Task 1: `stateFromCloudspace` mapping function

**Files:**
- Modify: `driver/config_test.go`
- Modify: `driver/config.go`

- [ ] **Step 1: Write the failing tests**

Append to `driver/config_test.go`:

```go
func TestStateFromCloudspace_fullConfig(t *testing.T) {
	p0 := &spotv1.SpotNodePool{
		Name:        "pool-a",
		ServerClass: "gp.vs1.medium-iad",
		Desired:     3,
		BidPrice:    "0.05",
	}
	p0.Autoscaling.Enabled  = true
	p0.Autoscaling.MinNodes = 2
	p0.Autoscaling.MaxNodes = 8

	p1 := &spotv1.SpotNodePool{
		Name:        "pool-b",
		ServerClass: "gp.vs1.large-iad",
		Desired:     2,
		BidPrice:    "0.08",
	}

	od := &spotv1.OnDemandNodePool{
		Name:                 "od-pool",
		ServerClass:          "gp.vs1.large-iad",
		Desired:              1,
		OnDemandPricePerHour: "0.50",
	}

	cs := &spotv1.CloudSpace{
		Name:                 "my-cluster",
		Region:               "us-east-iad-1",
		KubernetesVersion:    "1.33.0",
		CNI:                  "calico",
		GpuEnabled:           true,
		PreemptionWebhookURL: "https://example.com/webhook",
		DeploymentType:       "spot",
		SpotNodepools:        []*spotv1.SpotNodePool{p0, p1},
		OnDemandNodePools:    []*spotv1.OnDemandNodePool{od},
	}

	s := stateFromCloudspace(cs, "my-org", "tok")

	if s.Organization != "my-org" {
		t.Fatalf("Organization = %q, want my-org", s.Organization)
	}
	if s.RefreshToken != "tok" {
		t.Fatalf("RefreshToken = %q, want tok", s.RefreshToken)
	}
	if s.CloudspaceName != "my-cluster" {
		t.Fatalf("CloudspaceName = %q, want my-cluster", s.CloudspaceName)
	}
	if s.Region != "us-east-iad-1" {
		t.Fatalf("Region = %q, want us-east-iad-1", s.Region)
	}
	if s.KubernetesVersion != "1.33.0" {
		t.Fatalf("KubernetesVersion = %q, want 1.33.0", s.KubernetesVersion)
	}
	if s.CNI != "calico" {
		t.Fatalf("CNI = %q, want calico", s.CNI)
	}
	if !s.GPUEnabled {
		t.Fatal("GPUEnabled = false, want true")
	}
	if s.PreemptionWebhook != "https://example.com/webhook" {
		t.Fatalf("PreemptionWebhook = %q", s.PreemptionWebhook)
	}
	if s.DeploymentType != "spot" {
		t.Fatalf("DeploymentType = %q, want spot", s.DeploymentType)
	}
	if s.SpotPoolName != "pool-a" {
		t.Fatalf("SpotPoolName = %q, want pool-a", s.SpotPoolName)
	}
	if s.SpotServerClass != "gp.vs1.medium-iad" {
		t.Fatalf("SpotServerClass = %q", s.SpotServerClass)
	}
	if s.SpotNodeCount != 3 {
		t.Fatalf("SpotNodeCount = %d, want 3", s.SpotNodeCount)
	}
	if s.SpotBidPrice != "0.05" {
		t.Fatalf("SpotBidPrice = %q, want 0.05", s.SpotBidPrice)
	}
	if !s.SpotAutoscaling {
		t.Fatal("SpotAutoscaling = false, want true")
	}
	if s.SpotMinNodes != 2 {
		t.Fatalf("SpotMinNodes = %d, want 2", s.SpotMinNodes)
	}
	if s.SpotMaxNodes != 8 {
		t.Fatalf("SpotMaxNodes = %d, want 8", s.SpotMaxNodes)
	}
	if len(s.AdditionalSpotPools) != 1 {
		t.Fatalf("len(AdditionalSpotPools) = %d, want 1", len(s.AdditionalSpotPools))
	}
	p := s.AdditionalSpotPools[0]
	if p.Name != "pool-b" || p.ServerClass != "gp.vs1.large-iad" || p.NodeCount != 2 || p.BidPrice != "0.08" {
		t.Fatalf("AdditionalSpotPools[0] = %+v", p)
	}
	if !s.OnDemandEnabled {
		t.Fatal("OnDemandEnabled = false, want true")
	}
	if s.OnDemandPoolName != "od-pool" {
		t.Fatalf("OnDemandPoolName = %q, want od-pool", s.OnDemandPoolName)
	}
	if s.OnDemandClass != "gp.vs1.large-iad" {
		t.Fatalf("OnDemandClass = %q", s.OnDemandClass)
	}
	if s.OnDemandCount != 1 {
		t.Fatalf("OnDemandCount = %d, want 1", s.OnDemandCount)
	}
	if s.OnDemandPrice != "0.50" {
		t.Fatalf("OnDemandPrice = %q, want 0.50", s.OnDemandPrice)
	}
}

func TestStateFromCloudspace_noOnDemand(t *testing.T) {
	cs := &spotv1.CloudSpace{
		Name:          "cs",
		Region:        "us-east-iad-1",
		SpotNodepools: []*spotv1.SpotNodePool{{Name: "pool-a", Desired: 3}},
	}
	s := stateFromCloudspace(cs, "org", "tok")
	if s.OnDemandEnabled {
		t.Fatal("OnDemandEnabled = true, want false")
	}
	if s.OnDemandPoolName != "" || s.OnDemandClass != "" || s.OnDemandCount != 0 {
		t.Fatalf("unexpected on-demand fields: poolName=%q class=%q count=%d",
			s.OnDemandPoolName, s.OnDemandClass, s.OnDemandCount)
	}
}

func TestStateFromCloudspace_singleSpotPool(t *testing.T) {
	cs := &spotv1.CloudSpace{
		Name:          "cs",
		Region:        "us-east-iad-1",
		SpotNodepools: []*spotv1.SpotNodePool{{Name: "pool-a", Desired: 2}},
	}
	s := stateFromCloudspace(cs, "org", "tok")
	if s.SpotPoolName != "pool-a" {
		t.Fatalf("SpotPoolName = %q, want pool-a", s.SpotPoolName)
	}
	if len(s.AdditionalSpotPools) != 0 {
		t.Fatalf("len(AdditionalSpotPools) = %d, want 0", len(s.AdditionalSpotPools))
	}
}
```

The import at the top of `config_test.go` already has `"github.com/rancher/kontainer-engine/types"` and `"strings"` and `"testing"`. Add the Spot SDK import:

```go
import (
	"strings"
	"testing"

	spotv1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/rancher/kontainer-engine/types"
)
```

- [ ] **Step 2: Run the tests — expect compile error (stateFromCloudspace undefined)**

```bash
go test ./driver/ -run TestStateFromCloudspace -v 2>&1 | head -20
```

Expected: `./config_test.go:...: undefined: stateFromCloudspace`

- [ ] **Step 3: Implement `stateFromCloudspace` in `driver/config.go`**

Add this function after `stateFromClusterInfo` (around line 226):

```go
// stateFromCloudspace builds a clusterState by reading the live config of an
// existing cloudspace. org and token are taken from user-supplied opts since
// they are not stored on the cloudspace itself.
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

The import at the top of `config.go` already has `spotv1 "github.com/rackspace-spot/spot-go-sdk/api/v1"` — it does not currently; check the imports. If the SDK alias is missing, add it:

```go
import (
	"encoding/json"
	"fmt"
	"strings"

	spotv1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/google/uuid"
	"github.com/rancher/kontainer-engine/types"
	"github.com/sirupsen/logrus"
)
```

- [ ] **Step 4: Run the tests — expect PASS**

```bash
go test ./driver/ -run TestStateFromCloudspace -v
```

Expected:
```
--- PASS: TestStateFromCloudspace_fullConfig (0.00s)
--- PASS: TestStateFromCloudspace_noOnDemand (0.00s)
--- PASS: TestStateFromCloudspace_singleSpotPool (0.00s)
PASS
```

- [ ] **Step 5: Run the full driver test suite to check for regressions**

```bash
go test ./driver/ -v 2>&1 | tail -20
```

Expected: all existing tests pass.

- [ ] **Step 6: Commit**

```bash
git add driver/config.go driver/config_test.go
git commit -m "feat: add stateFromCloudspace mapping function"
```

---

## Task 2: `importCloudspace` on `spotClient`

**Files:**
- Modify: `driver/client.go`
- Create: `driver/import_test.go`

- [ ] **Step 1: Create `driver/import_test.go` with a mock and two failing tests**

```go
package driver

import (
	"context"
	"fmt"
	"strings"
	"testing"

	spotv1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
)

// importMockAPI is a minimal spotAPI implementation for import tests.
// Only GetCloudspace and CreateSpotNodePool are non-trivial.
type importMockAPI struct {
	cloudspace  *spotv1.CloudSpace
	notFound    bool
	createCalls []string
}

func (m *importMockAPI) Authenticate(_ context.Context) (string, error) { return "", nil }
func (m *importMockAPI) GetCloudspace(_ context.Context, _, _ string) (*spotv1.CloudSpace, error) {
	if m.notFound {
		return nil, fmt.Errorf("cloudspace not found")
	}
	return m.cloudspace, nil
}
func (m *importMockAPI) CreateCloudspace(_ context.Context, _ spotv1.CloudSpace) error { return nil }
func (m *importMockAPI) DeleteCloudspace(_ context.Context, _, _ string) error         { return nil }
func (m *importMockAPI) GetCloudspaceConfig(_ context.Context, _, _ string) (string, error) {
	return "", nil
}
func (m *importMockAPI) ListSpotNodePools(_ context.Context, _, _ string) ([]*spotv1.SpotNodePool, error) {
	return nil, nil
}
func (m *importMockAPI) CreateSpotNodePool(_ context.Context, _ string, p spotv1.SpotNodePool) error {
	m.createCalls = append(m.createCalls, p.Name)
	return nil
}
func (m *importMockAPI) UpdateSpotNodePool(_ context.Context, _ string, _ spotv1.SpotNodePool) error {
	return nil
}
func (m *importMockAPI) DeleteSpotNodePool(_ context.Context, _, _ string) error { return nil }
func (m *importMockAPI) GetOnDemandNodePool(_ context.Context, _, _ string) (*spotv1.OnDemandNodePool, error) {
	return nil, nil
}
func (m *importMockAPI) CreateOnDemandNodePool(_ context.Context, _ string, _ spotv1.OnDemandNodePool) error {
	return nil
}
func (m *importMockAPI) UpdateOnDemandNodePool(_ context.Context, _ string, _ spotv1.OnDemandNodePool) error {
	return nil
}
func (m *importMockAPI) ListOnDemandNodePools(_ context.Context, _, _ string) ([]*spotv1.OnDemandNodePool, error) {
	return nil, nil
}
func (m *importMockAPI) DeleteOnDemandNodePool(_ context.Context, _, _ string) error { return nil }

func TestImportCloudspace_notFound(t *testing.T) {
	client := &spotClient{api: &importMockAPI{notFound: true}, org: "org"}
	_, err := client.importCloudspace(context.Background(), "org", "missing-cluster", "tok")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error = %q, want it to contain 'not found'", err.Error())
	}
}

func TestImportCloudspace_success(t *testing.T) {
	cs := &spotv1.CloudSpace{
		Name:              "existing-cluster",
		Region:            "us-east-iad-1",
		KubernetesVersion: "1.33.0",
		CNI:               "calico",
		SpotNodepools: []*spotv1.SpotNodePool{
			{Name: "pool-x", ServerClass: "gp.vs1.medium-iad", Desired: 3, BidPrice: "0.05"},
		},
	}
	mock := &importMockAPI{cloudspace: cs}
	client := &spotClient{api: mock, org: "org"}

	s, err := client.importCloudspace(context.Background(), "org", "existing-cluster", "my-token")
	if err != nil {
		t.Fatalf("importCloudspace() error = %v", err)
	}
	if s.CloudspaceName != "existing-cluster" {
		t.Fatalf("CloudspaceName = %q, want existing-cluster", s.CloudspaceName)
	}
	if s.RefreshToken != "my-token" {
		t.Fatalf("RefreshToken = %q, want my-token", s.RefreshToken)
	}
	if s.SpotPoolName != "pool-x" {
		t.Fatalf("SpotPoolName = %q, want pool-x", s.SpotPoolName)
	}
	if len(mock.createCalls) != 0 {
		t.Fatalf("unexpected CreateSpotNodePool calls: %v", mock.createCalls)
	}
}
```

- [ ] **Step 2: Run the tests — expect compile error**

```bash
go test ./driver/ -run TestImportCloudspace -v 2>&1 | head -10
```

Expected: `undefined: (*spotClient).importCloudspace`

- [ ] **Step 3: Add `importCloudspace` to `driver/client.go`**

Append after `waitForCloudspace` (around line 274):

```go
// importCloudspace reads an existing cloudspace from the Spot API and returns
// its config as a clusterState. No infrastructure changes are made.
func (c *spotClient) importCloudspace(ctx context.Context, org, cloudspaceName, token string) (*clusterState, error) {
	cs, err := c.api.GetCloudspace(ctx, org, cloudspaceName)
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("cloudspace %q not found: use create mode to provision a new cluster", cloudspaceName)
		}
		return nil, fmt.Errorf("failed to look up cloudspace %q: %w", cloudspaceName, err)
	}
	return stateFromCloudspace(cs, org, token), nil
}
```

- [ ] **Step 4: Run the tests — expect PASS**

```bash
go test ./driver/ -run TestImportCloudspace -v
```

Expected:
```
--- PASS: TestImportCloudspace_notFound (0.00s)
--- PASS: TestImportCloudspace_success (0.00s)
PASS
```

- [ ] **Step 5: Run the full suite**

```bash
go test ./driver/ -v 2>&1 | tail -20
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
git add driver/client.go driver/import_test.go
git commit -m "feat: add importCloudspace method to spotClient"
```

---

## Task 3: `importCluster` driver method, `Create` branch, and flag declaration

**Files:**
- Modify: `driver/driver.go`

- [ ] **Step 1: Add `flagImportExisting` constant**

In `driver/config.go`, add to the `const` block after the other flag constants (around line 37):

```go
flagImportExisting = "import-existing-cluster"
```

- [ ] **Step 2: Add the flag to `GetDriverCreateOptions`**

In `driver/driver.go`, add to the `Options` map inside `GetDriverCreateOptions` (after `flagAdditionalSpotPools`, around line 134):

```go
flagImportExisting: {
    Type:    types.BoolType,
    Usage:   "Import an existing Rackspace Spot cloudspace instead of creating a new one",
    Default: &types.Default{DefaultBool: false},
},
```

- [ ] **Step 3: Add the import branch to `Create`**

In `driver/driver.go`, find the `Create` method. After the `logrus.Infof("[%s] Create() started", driverName)` line and before the `s, err := stateFromOptions(opts)` line, insert:

```go
if getBoolOption(opts, flagImportExisting, "importExistingCluster") {
    return d.importCluster(ctx, opts, clusterInfo)
}
```

- [ ] **Step 4: Add the `importCluster` method**

In `driver/driver.go`, add after the `Create` method (after line 270):

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

	logrus.Infof("[%s] importing cloudspace %q (org: %s)", driverName, cloudspaceName, org)

	client, err := newSpotClient(ctx, token, org)
	if err != nil {
		return nil, err
	}

	s, err := client.importCloudspace(ctx, org, cloudspaceName, token)
	if err != nil {
		return nil, err
	}

	info := clusterInfo
	if info == nil {
		info = &types.ClusterInfo{}
	}
	if err := s.save(info); err != nil {
		return info, err
	}

	logrus.Infof("[%s] imported cloudspace %s (region: %s, k8s: %s, spot pools: %d)",
		driverName, cloudspaceName, s.Region, s.KubernetesVersion, 1+len(s.AdditionalSpotPools))

	return info, nil
}
```

- [ ] **Step 5: Build to verify it compiles**

```bash
make build
```

Expected: `bin/kontainer-engine-driver-rackspacespot` produced with no errors.

- [ ] **Step 6: Run the full test suite**

```bash
go test ./...
```

Expected: all tests pass.

- [ ] **Step 7: Commit**

```bash
git add driver/config.go driver/driver.go
git commit -m "feat: add import-existing-cluster flag and importCluster driver method"
```

---

## Task 4: UI toggle and conditional pool-section hiding

**Files:**
- Modify: `ui/pkg/rackspacespot/components/CruRackspaceSpot.vue`

- [ ] **Step 1: Add `importExistingCluster` to `DEFAULTS`**

In `CruRackspaceSpot.vue`, find the `DEFAULTS` object (around line 307) and add:

```js
const DEFAULTS = {
  driverName:              DRIVER,
  rackspaceSpotRegion:     'us-east-iad-1',
  kubernetesVersion:       '1.33.0',
  cni:                     'calico',
  gpuEnabled:              false,
  importExistingCluster:   false,   // ← add this line
  spotServerClass:         'gp.vs1.medium-iad',
  // ... rest unchanged
};
```

- [ ] **Step 2: Add the import toggle to the template**

In the template, insert a new section between the Authentication section and the Cluster `<h3>` (after the closing `</div>` of the authentication row, around line 51):

```html
    <!-- ── Import mode ───────────────────────────────────── -->
    <div class="row mt-20">
      <div class="col span-12">
        <Checkbox
          v-model:value="config.importExistingCluster"
          label="Import existing cluster"
          :mode="mode === 'edit' ? 'view' : mode"
        />
        <p v-if="config.importExistingCluster" class="import-note mt-5">
          Node pool configuration will be read from the existing cloudspace after import.
        </p>
      </div>
    </div>
```

- [ ] **Step 3: Wrap pool sections in `v-if`**

In the template, find the `<!-- ── Spot Node Pools ───────────────────────────────── -->` comment (around line 98). Wrap everything from that comment through the closing `</template>` of the on-demand section in a single `<template v-if>`:

Replace from `<!-- ── Spot Node Pools -->` comment to the last `</template>` before `</CruResource>` with:

```html
    <!-- ── Node pools (hidden in import mode) ───────────── -->
    <template v-if="!config.importExistingCluster">
      <!-- ── Spot Node Pools ───────────────────────────────── -->
      <h3 class="mt-20">Spot Node Pools</h3>
      <a href="https://tombojer.github.io/spot-cost-analyzer/" target="_blank" rel="noopener" class="spot-cost-link">Estimate costs with the Spot Cost Analyzer ↗</a>

      <!-- ── Primary spot pool ─────────────────────────────── -->
      <div class="pool-card mt-15">
        <div class="pool-card-header">
          <span class="pool-label">Pool 1 (primary)</span>
        </div>
        <div class="row mt-10">
          <div class="col span-4">
            <LabeledSelect
              v-model:value="config.spotServerClass"
              label="Server Class"
              :options="serverClassOptions"
              :taggable="true"
              :mode="mode"
            />
          </div>
          <div class="col span-4">
            <LabeledInput
              v-model:value="config.spotNodeCount"
              label="Node Count"
              type="number"
              :min="0"
              :mode="mode"
            />
          </div>
          <div class="col span-4">
            <LabeledInput
              v-model:value="config.spotBidPrice"
              label="Bid Price (USD/hr)"
              placeholder="0.01"
              :mode="mode"
            />
          </div>
        </div>
        <div class="row mt-10">
          <div class="col span-4">
            <Checkbox
              v-model:value="config.spotAutoscalingEnabled"
              :label="primaryAutoscalingDisabled ? 'Enable Autoscaling (another pool has autoscaling)' : 'Enable Autoscaling'"
              :disabled="primaryAutoscalingDisabled"
              :mode="mode"
            />
          </div>
          <template v-if="config.spotAutoscalingEnabled">
            <div class="col span-4">
              <LabeledInput
                v-model:value="config.spotAutoscalingMinNodes"
                label="Min Nodes"
                type="number"
                :min="0"
                :mode="mode"
              />
            </div>
            <div class="col span-4">
              <LabeledInput
                v-model:value="config.spotAutoscalingMaxNodes"
                label="Max Nodes"
                type="number"
                :min="1"
                :mode="mode"
              />
            </div>
          </template>
        </div>
      </div>

      <!-- ── Additional spot pools ─────────────────────────── -->
      <div
        v-for="(pool, idx) in additionalSpotPools"
        :key="idx"
        class="pool-card mt-10"
      >
        <div class="pool-card-header">
          <span class="pool-label">Pool {{ idx + 2 }}</span>
          <button
            class="btn btn-sm btn-danger"
            type="button"
            @click="removeSpotPool(idx)"
          >
            Remove
          </button>
        </div>
        <div class="row mt-10">
          <div class="col span-4">
            <LabeledSelect
              v-model:value="pool.serverClass"
              label="Server Class"
              :options="serverClassOptions"
              :taggable="true"
              :mode="mode"
            />
          </div>
          <div class="col span-4">
            <LabeledInput
              v-model:value="pool.nodeCount"
              label="Node Count"
              type="number"
              :min="0"
              :mode="mode"
            />
          </div>
          <div class="col span-4">
            <LabeledInput
              v-model:value="pool.bidPrice"
              label="Bid Price (USD/hr)"
              placeholder="0.01"
              :mode="mode"
            />
          </div>
        </div>
        <div class="row mt-10">
          <div class="col span-4">
            <Checkbox
              v-model:value="pool.autoscaling"
              :label="autoscalingPoolCount >= 1 && !pool.autoscaling ? 'Enable Autoscaling (another pool has autoscaling)' : 'Enable Autoscaling'"
              :disabled="autoscalingPoolCount >= 1 && !pool.autoscaling"
              :mode="mode"
            />
          </div>
          <template v-if="pool.autoscaling">
            <div class="col span-4">
              <LabeledInput
                v-model:value="pool.minNodes"
                label="Min Nodes"
                type="number"
                :min="0"
                :mode="mode"
              />
            </div>
            <div class="col span-4">
              <LabeledInput
                v-model:value="pool.maxNodes"
                label="Max Nodes"
                type="number"
                :min="1"
                :mode="mode"
              />
            </div>
          </template>
        </div>
      </div>

      <div class="mt-10">
        <button
          class="btn btn-sm btn-primary"
          type="button"
          @click="addSpotPool"
        >
          + Add Spot Pool
        </button>
      </div>

      <!-- ── On-Demand Node Pool ─────────────────────────────── -->
      <h3 class="mt-20">On-Demand Node Pool</h3>
      <div class="row">
        <div class="col span-12">
          <Checkbox
            v-model:value="config.onDemandEnabled"
            label="Add stable capacity alongside spot nodes"
            :mode="mode"
          />
        </div>
      </div>
      <template v-if="config.onDemandEnabled">
        <div class="row mt-10">
          <div class="col span-4">
            <LabeledSelect
              v-model:value="config.onDemandServerClass"
              label="Server Class"
              :options="serverClassOptions"
              :taggable="true"
              :mode="mode"
            />
          </div>
          <div class="col span-4">
            <LabeledInput
              v-model:value="config.onDemandNodeCount"
              label="Node Count"
              type="number"
              :min="0"
              :mode="mode"
            />
          </div>
          <div class="col span-4">
            <LabeledInput
              v-model:value="config.onDemandPricePerHour"
              label="Price Per Hour (USD)"
              placeholder="0.00"
              :mode="mode"
            />
          </div>
        </div>
      </template>
    </template>
```

- [ ] **Step 4: Skip pool validation in `save()` when in import mode**

In `save()`, find the two pool-related validation blocks (around line 509):

```js
      if (this.config.onDemandEnabled && !this.config.onDemandServerClass) {
        this.errors = ['On-Demand Server Class is required when the on-demand pool is enabled'];
        if (btnCb) btnCb(false);
        return;
      }
      if (this.autoscalingPoolCount > 1) {
        this.errors = ['Only one spot node pool may have autoscaling enabled per cloudspace (API limit).'];
        if (btnCb) btnCb(false);
        return;
      }
```

Wrap both in a single `if (!this.config.importExistingCluster)` block:

```js
      if (!this.config.importExistingCluster) {
        if (this.config.onDemandEnabled && !this.config.onDemandServerClass) {
          this.errors = ['On-Demand Server Class is required when the on-demand pool is enabled'];
          if (btnCb) btnCb(false);
          return;
        }
        if (this.autoscalingPoolCount > 1) {
          this.errors = ['Only one spot node pool may have autoscaling enabled per cloudspace (API limit).'];
          if (btnCb) btnCb(false);
          return;
        }
      }
```

- [ ] **Step 5: Add `.import-note` style**

In the `<style scoped>` section, add:

```css
.import-note {
  font-size: 0.875rem;
  color: var(--muted);
}
```

- [ ] **Step 6: Build the UI extension**

```bash
cd ui && npm run build-pkg rackspacespot 2>&1 | tail -10
```

Expected: build completes with no errors; output in `ui/dist-pkg/rackspacespot-*/`.

- [ ] **Step 7: Commit**

```bash
git add ui/pkg/rackspacespot/components/CruRackspaceSpot.vue
git commit -m "feat(ui): add import-existing-cluster toggle with conditional pool sections"
```

---

## Task 5: UI test for import-mode validation bypass

**Files:**
- Modify: `ui/pkg/rackspacespot/test.mjs`

- [ ] **Step 1: Add `validateConfig` function and tests to `test.mjs`**

Append to `ui/pkg/rackspacespot/test.mjs` (before the final `console.log` summary):

```js
// Mirror of the save() validation logic in CruRackspaceSpot.vue.
// Returns an error message string, or null if valid.
function validateConfig(clusterName, config, autoscalingPoolCount) {
  if (!clusterName) return 'Cluster Name is required';
  if (!config.rackspaceSpotRefreshToken) return 'Refresh Token is required';
  if (!config.rackspaceSpotOrganization) return 'Organization is required';
  if (!config.importExistingCluster) {
    if (config.onDemandEnabled && !config.onDemandServerClass) {
      return 'On-Demand Server Class is required when the on-demand pool is enabled';
    }
    if (autoscalingPoolCount > 1) {
      return 'Only one spot node pool may have autoscaling enabled per cloudspace (API limit).';
    }
  }
  return null;
}

console.log('\nimport mode validation:');
test('import mode skips on-demand server class check', () => {
  const err = validateConfig('my-cluster', {
    rackspaceSpotRefreshToken: 'tok',
    rackspaceSpotOrganization: 'org',
    onDemandEnabled:           true,
    onDemandServerClass:       '',
    importExistingCluster:     true,
  }, 0);
  assert.equal(err, null);
});
test('normal mode requires on-demand server class when enabled', () => {
  const err = validateConfig('my-cluster', {
    rackspaceSpotRefreshToken: 'tok',
    rackspaceSpotOrganization: 'org',
    onDemandEnabled:           true,
    onDemandServerClass:       '',
    importExistingCluster:     false,
  }, 0);
  assert.ok(err !== null && err.includes('On-Demand'), `expected On-Demand error, got: ${err}`);
});
test('import mode skips autoscaling pool count check', () => {
  const err = validateConfig('my-cluster', {
    rackspaceSpotRefreshToken: 'tok',
    rackspaceSpotOrganization: 'org',
    importExistingCluster:     true,
  }, 2);
  assert.equal(err, null);
});
test('normal mode rejects multiple autoscaling pools', () => {
  const err = validateConfig('my-cluster', {
    rackspaceSpotRefreshToken: 'tok',
    rackspaceSpotOrganization: 'org',
    importExistingCluster:     false,
  }, 2);
  assert.ok(err !== null && err.includes('autoscaling'), `expected autoscaling error, got: ${err}`);
});
```

The `validateConfig` function and the pool sections in `CruRackspaceSpot.vue` use the same logic — the test validates that `importExistingCluster: true` skips both pool-related validation checks.

- [ ] **Step 2: Run the UI tests**

```bash
cd ui && npm test
```

Expected:
```
import mode validation:
  ✓ import mode skips on-demand server class check
  ✓ normal mode requires on-demand server class when enabled
  ✓ import mode skips autoscaling pool count check
  ✓ normal mode rejects multiple autoscaling pools

N passed, 0 failed
```

- [ ] **Step 3: Commit**

```bash
git add ui/pkg/rackspacespot/test.mjs
git commit -m "test(ui): add import mode validation tests"
```

---

## Final check

- [ ] **Run the full Go test suite one last time**

```bash
go test ./...
```

Expected: all tests pass, no failures.

- [ ] **Run the UI tests one last time**

```bash
cd ui && npm test
```

Expected: all tests pass, 0 failed.
