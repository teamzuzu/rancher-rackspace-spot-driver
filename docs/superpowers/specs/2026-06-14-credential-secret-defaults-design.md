# Credential Secret Defaults

**Date:** 2026-06-14  
**Status:** Approved

## Problem

Every Rackspace Spot cluster created via Rancher requires the user to enter an org name and refresh token in the UI form. In environments with a single Rackspace Spot organisation, this is repetitive and error-prone. There is no mechanism to pre-configure these credentials at the driver level.

## Goal

Allow an operator to create a single Kubernetes Secret containing the org name and refresh token. The driver reads this secret and uses the values as defaults in the Rancher create-cluster form. Users see the fields pre-filled and can override them; clusters can also be provisioned programmatically without providing credentials in the UI at all.

## Architecture

The driver binary runs as a subprocess of Rancher's pod and inherits its in-cluster service account. This gives the driver access to the Kubernetes API without any additional configuration.

```
Admin creates Secret  →  Helm chart creates RBAC  →  Driver reads Secret
      ↓                                                       ↓
rackspace-spot-credentials              GetDriverCreateOptions fills defaults
in cattle-system                        stateFromOptions falls back if UI blank
```

**Secret discovery:** The driver always looks for a Secret named `rackspace-spot-credentials` in namespace `cattle-system`. No configuration is required — if the secret exists the feature activates, if it does not the existing behaviour is unchanged.

**Precedence:** UI value wins. The secret provides the default shown in the form. If the user clears the field and submits, `stateFromOptions` fills it in from the secret before validation runs (handles the case where Rancher does not echo back password field defaults).

**Graceful degradation:** `loadCredentialDefaults` never returns an error. If the driver is running outside a cluster, the secret is missing, or RBAC is denied, it logs a warning and returns empty defaults. Existing validation (`rackspace-spot-refresh-token is required`) still fires if the user also left the field blank.

## Components

### `driver/secrets.go` (new file)

```go
type credentialDefaults struct {
    Org          string
    RefreshToken string
}

func loadCredentialDefaults(ctx context.Context) credentialDefaults
```

- Calls `rest.InClusterConfig()`. Returns empty struct immediately if this fails (not running in-cluster).
- Reads Secret `rackspace-spot-credentials` in namespace `cattle-system`.
- Extracts keys `org` and `refreshToken`. Missing keys return an empty string for that field only.
- On any error (not found, RBAC denied, etc.) logs at Warn level and returns empty struct.

### `driver/driver.go` — `GetDriverCreateOptions`

After building the flags map, call `loadCredentialDefaults(ctx)`. For any non-empty value returned, set `Default.DefaultString` on the corresponding flag (`flagOrganization`, `flagRefreshToken`).

### `driver/config.go` — `stateFromOptions`

After populating `s` from opts, if `s.Organization == ""` or `s.RefreshToken == ""`, call `loadCredentialDefaults(ctx)` and fill in the missing fields. This runs before `applyDefaults` and `validate`.

`stateFromOptions` gains a `ctx context.Context` parameter (currently has none). Call sites in `driver.go` pass through their existing `ctx`.

### `deploy/charts/rackspacespot/values.yaml`

```yaml
credentials:
  secretName: "rackspace-spot-credentials"
  namespace: "cattle-system"
  orgKey: "org"
  refreshTokenKey: "refreshToken"
```

The key names in `values.yaml` document the expected secret structure. They are not read by the driver at runtime (the driver uses hardcoded key names matching the defaults); they serve as documentation and drive the RBAC template.

### `deploy/charts/rackspacespot/templates/credentials-rbac.yaml` (new file)

Rendered only when `credentials.secretName` is non-empty. Creates:

- `Role` in `credentials.namespace`: `get` verb on `secrets` resource, `resourceNames` restricted to `credentials.secretName`.
- `RoleBinding`: binds the Role to the Rancher ServiceAccount (`cattle-system/rancher`).

## Secret Format

Operators create the secret manually before deploying the chart:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: rackspace-spot-credentials
  namespace: cattle-system
type: Opaque
stringData:
  org: "my-rackspace-org"
  refreshToken: "eyJ..."
```

## Error Handling

| Scenario | Behaviour |
|---|---|
| Not running in-cluster (e.g. tests) | `InClusterConfig` fails → empty defaults, no error |
| Secret does not exist | 404 → warn log, empty defaults |
| Secret exists, key missing | Empty string for that field only |
| RBAC denied | Error reading secret → warn log, empty defaults |
| Both secret and UI blank | Existing `validate()` returns "rackspace-spot-refresh-token is required" |

## Testing

**`driver/secrets_test.go`** (new file):
- Secret found with both keys → both fields populated
- Secret found, `org` key missing → only `RefreshToken` populated
- Secret not found (404) → empty defaults, no error returned
- `InClusterConfig` fails → returns immediately with empty defaults

**`driver/config_test.go`** additions:
- `stateFromOptions` with blank org/token + mock secret → fields filled from secret
- `stateFromOptions` with explicit org/token → secret values not used (UI wins)
- `stateFromOptions` with no secret + blank fields → validation error unchanged

## Out of Scope

- Configurable secret name/namespace (well-known name is sufficient for the target use case)
- UI changes — the form pre-fill is handled entirely by `GetDriverCreateOptions` returning a `Default`

## Notes

- Secret rotation takes effect on the next cluster operation without a pod restart, because the driver reads the secret per-call.
