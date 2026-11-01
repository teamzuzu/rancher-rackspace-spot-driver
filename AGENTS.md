# Repository Guidelines

## Project Structure & Module Organization

This repository pairs a Go Rancher KontainerDriver with a Vue 3 Dashboard extension for Rackspace Spot.

- `main.go` starts the driver; `driver/` contains lifecycle operations, API reconciliation, configuration, and Go tests.
- `cmd/inspect-cloudspace/` provides a diagnostic CLI.
- `ui/pkg/rackspacespot/` contains extension registration, the provisioner form, SVG assets, and JavaScript tests.
- `deploy/charts/` contains Helm packaging; `docs/` contains site content and published extension artifacts.
- `.github/workflows/` defines CI and release automation; `bin/` receives local builds.

## Build, Test, and Development Commands

Use Go 1.25 or later and Node.js 20 or later.

- `make build`: build `bin/rancher-rackspace-spot-driver` with version metadata.
- `make test`: run all Go tests.
- `make vet`: run Go static checks.
- `make lint`: run installed golangci-lint using `.golangci.yml`.
- `make tidy`: reconcile Go dependencies.
- From `ui/`, run `npm install --legacy-peer-deps`, then `npm run build-pkg rackspacespot` to build the extension into `dist-pkg/`.
- From `ui/`, run `npm test` for JavaScript checks or `npm run dev` for local development with a running Rancher instance.

## Coding Style & Naming Conventions

Format Go with `gofmt` or `go fmt ./...`; use tabs and idiomatic Go names. Keep implementation files in their existing packages. Follow the UI’s two-space indentation, camelCase properties, and PascalCase Vue component filenames. The Go lint configuration enables checks including errcheck, revive, and staticcheck.

## Testing Guidelines

Use Go’s `testing` package with `*_test.go` files and descriptive `Test...` functions. Cover configuration validation, token redaction, import behavior, and pool reconciliation when changing those paths. UI tests use Node’s strict assertions in `ui/pkg/rackspacespot/test.mjs`, without a browser. Run both suites for changes affecting driver/UI configuration contracts. No numeric coverage threshold is configured.

## Commit & Pull Request Guidelines

Follow the history’s conventional prefixes, such as `fix(driver): ...`, `chore: ...`, and `docs: ...`. Keep commits focused. PRs should describe behavior changes, link relevant issues, and report validation commands; include screenshots for visible form changes.

## Security & Compatibility

Keep refresh tokens out of logs and serialized metadata; driver credentials belong in `ClusterInfo.Password`. Preserve the Kubernetes v0.27.4 replacement pins in `go.mod`: older kontainer-engine dependencies require APIs removed in later versions.
