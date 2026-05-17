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

**Spot pool reconciliation** (`client.go:reconcileSpotNodePools`) runs in three phases to respect the Rackspace constraint of at most one autoscaling pool per cloudspace: (1) delete removed pools, (2) disable autoscaling on pools losing it, (3) create/update remaining pools.

**Token handling** (`driver/util.go`): creates a `rancher` service account in `kube-system` with cluster-admin, then waits up to 30s for the token Secret — required because k8s 1.24+ no longer auto-populates token secrets.

**k8s version pin**: `go.mod` forces `k8s.io/{api,apimachinery,client-go}` to v0.27.4 via `replace` directives. Do not upgrade these — `kontainer-engine` transitively imports alpha API packages removed in v0.28+.

### UI extension (`ui/`)

A Vue 3 plugin built with `@rancher/shell` 3.0.8 that adds the provisioner form to Rancher Dashboard.

- `ui/pkg/rackspacespot/index.ts` — registers the provisioner plugin and i18n
- `ui/pkg/rackspacespot/provisioner.ts` — implements `IClusterProvisioner`; defines the label, inline-SVG icon, and component reference
- `ui/pkg/rackspacespot/components/CruRackspaceSpot.vue` — the entire form (~500 lines); handles auth, cluster config, primary spot pool, optional on-demand pool, additional spot pools

**Server class naming**: Class names are region-specific (e.g. `gp.vs1.large-iad`). The component's `serverClassOptions` computed property derives a location code from the region string (second-to-last hyphen segment: `us-east-iad-1` → `iad`, `aus-syd-1` → `syd`) and substitutes it into the base `-iad` names. The test file `ui/pkg/rackspacespot/test.mjs` validates this logic without requiring a browser or framework.

**UI build output** goes to `ui/dist-pkg/rackspacespot-<version>/`. The version comes from `ui/pkg/rackspacespot/package.json`. After building, the extension bundle URL in Rancher is version-specific — a version bump requires `helm upgrade` to register the new URL.

### CI / Release

- `.github/workflows/ci.yml` — runs on PRs; all jobs use `container: ghcr.io/${{ github.repository }}-ci:latest`
- `Dockerfile.ci` — the CI container; bump `ARG` versions here to upgrade tools; triggers `docker-ci.yml` rebuild on push to main
- `.github/workflows/release.yml` — triggered by semver tags; builds binaries via GoReleaser, packages the UI extension and Helm chart, updates `docs/index.yaml` and commits back to main
- `deploy/charts/` — Helm chart for installing the driver into a Rancher management cluster

The release workflow calls `gh release upload` and `gh workflow run` — these require the `gh` CLI which is installed in `Dockerfile.ci`.
