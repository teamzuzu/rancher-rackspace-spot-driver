# Development

This guide covers building, testing, and releasing the driver locally.

---

## Prerequisites

| Tool | Version | Notes |
|---|---|---|
| Go | ≥ 1.24 | `go.mod` specifies the minimum |
| Docker (optional) | Any | For building the container image |
| golangci-lint (optional) | Latest | Or rely on the CI lint job |

---

## Building

```bash
# Clone the repository
git clone https://github.com/teamzuzu/rancher-rackspace-spot-driver.git
cd rancher-rackspace-spot-driver

# Sync dependencies
go mod tidy

# Build for your current platform
go build -o bin/kontainer-engine-driver-rackspacespot .

# Cross-compile for linux/amd64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -ldflags "-s -w" -o bin/kontainer-engine-driver-rackspacespot .

# Cross-compile for linux/arm64
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
  go build -ldflags "-s -w" -o bin/kontainer-engine-driver-rackspacespot-arm64 .
```

---

## Running tests

```bash
go test -v -race -count=1 ./...
```

Tests use the standard library only — no external test dependencies are required.

---

## Linting

```bash
# Format
go fmt ./...

# Vet
go vet ./...

# Full lint (requires golangci-lint)
golangci-lint run --timeout 5m
```

The CI pipeline runs all three checks on every pull request.

---

## Project layout

```
.
├── main.go                  # Entry point — starts the gRPC driver server
├── driver/
│   ├── driver.go            # types.Driver interface implementation
│   ├── client.go            # Rackspace Spot API client wrapper
│   ├── config.go            # Cluster state, flags, and serialization
│   └── util.go              # Kubeconfig parsing + Rancher service account bootstrap
├── docs/                    # MkDocs source (this site)
├── .github/
│   └── workflows/
│       ├── ci.yml           # Lint, build, test on every PR
│       ├── release.yml      # Manual release workflow
│       └── pages.yml        # GitHub Pages deployment
├── .golangci.yml            # golangci-lint configuration
├── .goreleaser.yml          # GoReleaser build configuration
└── mkdocs.yml               # MkDocs site configuration
```

---

## Making changes

1. **Fork and clone** the repository
2. **Create a branch**: `git checkout -b feature/my-change`
3. **Make changes** and add tests where appropriate
4. **Run CI checks locally**:
   ```bash
   go mod tidy && go fmt ./... && go vet ./... && go test -race ./...
   ```
5. **Open a pull request** against `main`

All PRs must pass the CI pipeline before merge. The pipeline runs:

- `golangci-lint` (format, errcheck, staticcheck, and more)
- `go mod tidy` drift check
- `go build` for linux/amd64 and linux/arm64
- `go test -race`
- `go vet`

---

## Architecture notes

### Driver lifecycle

Rancher calls the driver over a local gRPC socket. The sequence for cluster creation is:

```
Rancher → Create()     → provisions CloudSpace + node pools
        → PostCheck()  → waits for Ready, fetches kubeconfig, bootstraps service account
```

For updates:

```
Rancher → Update()     → reconciles node pools (create/update as needed)
        → PostCheck()  → re-validates readiness
```

For deletion:

```
Rancher → Remove()     → deletes all node pools, then deletes the CloudSpace
```

### State persistence

The driver serializes cluster state (pool config, CloudSpace name, organization) into `ClusterInfo.Metadata["state"]` as JSON. The refresh token is stored separately in `ClusterInfo.Password` so it is never included in the logged metadata blob. Rancher stores both and passes them back on every subsequent call. There is no external state store.

### k8s dependency pinning

The `go.mod` contains `replace` directives that pin `k8s.io/api`, `k8s.io/apimachinery`, and `k8s.io/client-go` to `v0.27.4`. This is required because `kontainer-engine` transitively depends on `rancher/rke`, which imports `k8s.io/client-go v12.0.0+incompatible`. That version references alpha API packages (`auditregistration/v1alpha1`, `batch/v2alpha1`) that were removed in k8s.io/api ≥ v0.28. The pin keeps those packages available without breaking the module graph.

---

## Releasing

Releases are created via the **Release** workflow in GitHub Actions — no manual tagging needed.

1. Go to **Actions → Release → Run workflow**
2. Enter the version in `vMAJOR.MINOR.PATCH` format (e.g. `v1.2.3`)
3. Optionally check **Publish as draft release** to review before it goes public
4. Click **Run workflow**

The workflow will:

- Validate the version format
- Create and push a signed git tag
- Run GoReleaser to build linux/amd64 and linux/arm64 binaries and publish a GitHub Release
- Build and push a multi-arch container image to `ghcr.io/teamzuzu/rancher-rackspace-spot-driver` (the image contains the `kontainer-engine-driver-rackspacespot` binary)

!!! note "Permissions"
    Only repository maintainers with **write** access can trigger the release workflow.

---

## Contributing

Issues and pull requests are welcome. Please open an issue first for significant changes so the approach can be discussed before you invest time in an implementation.
