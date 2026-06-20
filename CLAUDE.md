# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Go driver
make build          # Build binary → bin/kontainer-engine-driver-rackspacespot
make test           # go test ./...
make lint           # golangci-lint run ./...
make vet            # go vet ./...
make tidy           # go mod tidy

# UI extension (run from ui/)
npm install --legacy-peer-deps
npm run build-pkg rackspacespot   # Build → ui/dist-pkg/rackspacespot-<version>/
npm test                          # node pkg/rackspacespot/test.mjs (pure Node, no browser)
npm run dev                       # Local dev server (requires Rancher running)
```

## Architecture

This repo is a **Rancher KontainerDriver** that provisions Kubernetes clusters on the [Rackspace Spot](https://spot.rackspace.com) platform. It has two independently-deployed pieces:

### Driver binary (`main.go`, `driver/`)

The binary exposes a gRPC server that Rancher calls for cluster lifecycle operations (Create, Update, Remove, PostCheck, etc.) via the `rancher/kontainer-engine` interface.

**Stateless design**: The driver stores all cluster state as JSON in `ClusterInfo.Metadata` and the refresh token in `ClusterInfo.Password`. Each gRPC call deserializes state from these fields via `stateFromOptions()` in `driver/config.go`.

Key files:
- `driver/driver.go` — implements `types.Driver`; entry points for all Rancher-initiated operations
- `driver/client.go` — wraps `spot-go-sdk`; handles cloudspace and node pool reconciliation
- `driver/config.go` — `clusterState` struct, option parsing, validation, JSON serialization
- `driver/rancher.go` — `syncGenericEngineConfig`: patches `spec.genericEngineConfig` on the Rancher cluster object via in-cluster k8s API after a successful import, so the edit form reflects actual pool state instead of create-form defaults
- `driver/util.go` — kubeconfig parsing, Rancher service account creation

**Spot pool reconciliation** (`client.go:reconcileSpotNodePools`) runs in three phases to respect the Rackspace constraint of at most one autoscaling pool per cloudspace: (1) delete removed pools, (2) disable autoscaling on pools losing it, (3) create/update remaining pools.

**Import existing cloudspace** (`driver.go:importCluster`): when `importExistingCluster=true`, fetches the live cloudspace via the Spot API and builds `clusterState` from it (`stateFromCloudspace` in `config.go`). `Remove()` skips cloudspace deletion for imported clusters (`Imported: true` in state). After `PostCheck()` completes, `syncGenericEngineConfig` patches the Rancher cluster CRD so the edit form shows actual pool values (the create form submits defaults for hidden pool fields). The Rancher cluster ID is stored in `clusterState.RancherClusterID` for this purpose.

**Bid price normalization**: The Spot SDK's `ListSpotNodePools` prepends `$` to bid prices (e.g. `"$0.001"`), but the API rejects that prefix on writes. `normalizeBidPrice()` in `config.go` strips it; called in `stateFromCloudspace` and `syncGenericEngineConfig`.

**Token handling** (`driver/util.go`): creates a `rancher` service account in `kube-system` with cluster-admin, then waits up to 30s for the token Secret — required because k8s 1.24+ no longer auto-populates token secrets.

**k8s version pin**: `go.mod` forces `k8s.io/{api,apimachinery,client-go}` to v0.27.4 via `replace` directives. Do not upgrade these — `kontainer-engine` transitively imports alpha API packages removed in v0.28+.

### UI extension (`ui/`)

A Vue 3 plugin built with `@rancher/shell` 3.0.8 that adds the provisioner form to Rancher Dashboard.

- `ui/pkg/rackspacespot/index.ts` — registers the provisioner plugin and i18n
- `ui/pkg/rackspacespot/provisioner.ts` — implements `IClusterProvisioner`; defines the label, inline-SVG icon, and component reference
- `ui/pkg/rackspacespot/components/CruRackspaceSpot.vue` — the entire form (~500 lines); handles auth, cluster config, primary spot pool, optional on-demand pool, additional spot pools, import toggle

**Import mode (UI)**: `importExistingCluster` toggle is locked to view-only in edit mode (one-time create action). When active, the cluster config section and all pool sections are hidden (`v-if="!config.importExistingCluster || mode === 'edit'"`). A dedicated `importCloudspaceName` field is shown — this must be the exact cloudspace name, not the Rancher display name. In edit mode, pool sections are always shown regardless of import state so the user can manage pools after import.

**Server class naming**: Class names are region-specific (e.g. `gp.vs1.large-iad`). The component's `serverClassOptions` computed property derives a location code from the region string (second-to-last hyphen segment: `us-east-iad-1` → `iad`, `aus-syd-1` → `syd`) and substitutes it into the base `-iad` names. The test file `ui/pkg/rackspacespot/test.mjs` validates this logic without requiring a browser or framework.

**UI build output** goes to `ui/dist-pkg/rackspacespot-<version>/`. The version comes from `ui/pkg/rackspacespot/package.json`. After building, the extension bundle URL in Rancher is version-specific — a version bump requires `helm upgrade` to register the new URL.

### CI / Release

- `.github/workflows/ci.yml` — runs on PRs; all jobs use `container: ghcr.io/${{ github.repository }}-ci:latest`
- `Dockerfile.ci` — the CI container; bump `ARG` versions here to upgrade tools; triggers `docker-ci.yml` rebuild on push to main
- `.github/workflows/auto-release.yml` — triggers on source file changes to main; computes next patch version from latest git tag and dispatches `release.yml`
- `.github/workflows/release.yml` — triggered by semver tags or `workflow_dispatch`; builds binaries via GoReleaser, packages the UI extension and Helm chart, updates `docs/index.yaml` and commits back to main
- `deploy/charts/` — Helm chart for installing the driver into a Rancher management cluster

The release workflow calls `gh release upload` and `gh workflow run` — these require the `gh` CLI which is installed in `Dockerfile.ci`.

**Release flow**: pushing source changes to main triggers `auto-release.yml`, which computes the next patch version and dispatches `release.yml`. You can also push a semver tag directly to trigger `release.yml` — but this causes a duplicate run (tag-push + the auto-release dispatch both fire); the second GoReleaser run fails with "already_exists" which is harmless since the first run completes successfully.

**Upgrading the live cluster** after a release:
```bash
helm repo update rackspace-spot
helm upgrade rackspacespot rackspace-spot/rackspacespot -n cattle-ui-plugin-system
```
