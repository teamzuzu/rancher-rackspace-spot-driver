# Credential Secret Defaults Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow an operator to create a Kubernetes Secret with `org` and `refreshToken` keys; the driver reads it at runtime and uses the values as defaults in the Rancher cluster-create form.

**Architecture:** A new `driver/secrets.go` introduces `loadCredentialDefaultsFromClient` (testable, takes `kubernetes.Interface`) and `loadCredentialDefaults` (production, uses in-cluster config). A package-level `defaultCredentialLoader` variable lets tests swap the loader without real k8s access. `stateFromOptions` gains a `ctx` parameter and calls the loader as a fallback when UI fields are blank. `GetDriverCreateOptions` also calls it to pre-fill form defaults. The Helm chart gains a `credentials-rbac.yaml` template that creates a least-privilege `Role`+`RoleBinding` so Rancher's SA can read the named secret.

**Tech Stack:** Go 1.25, `k8s.io/client-go v0.27.4` (`kubernetes`, `kubernetes/fake`, `rest`), `k8s.io/api/core/v1`, `k8s.io/apimachinery`, Helm 3

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `driver/secrets.go` | Create | `credentialDefaults` struct, constants, `loadCredentialDefaultsFromClient`, `loadCredentialDefaults`, `defaultCredentialLoader` |
| `driver/secrets_test.go` | Create | Unit tests for `loadCredentialDefaultsFromClient` using `fake.NewSimpleClientset` |
| `driver/config.go` | Modify | Add `ctx` to `stateFromOptions`; call `defaultCredentialLoader` before `validate` |
| `driver/config_test.go` | Modify | Add `context.Background()` to existing call; three new tests using swapped loader |
| `driver/driver.go` | Modify | Pass `ctx` to `stateFromOptions` in `Create`; inject defaults in `GetDriverCreateOptions` |
| `driver/driver_test.go` | Create | Two tests for `GetDriverCreateOptions` defaults via swapped loader |
| `deploy/charts/rackspacespot/values.yaml` | Modify | Add `credentials` section |
| `deploy/charts/rackspacespot/templates/credentials-rbac.yaml` | Create | `Role` + `RoleBinding` rendered when `credentials.secretName` is non-empty |

---

## Task 1: Secret-reading core (`driver/secrets.go` + `driver/secrets_test.go`)

**Files:**
- Create: `driver/secrets.go`
- Create: `driver/secrets_test.go`

- [ ] **Step 1: Write the failing tests**

Create `driver/secrets_test.go`:

```go
package driver

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLoadCredentialDefaultsFromClientBothKeys(t *testing.T) {
	client := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      credentialsSecretName,
			Namespace: credentialsNamespace,
		},
		Data: map[string][]byte{
			credentialsOrgKey:   []byte("my-org"),
			credentialsTokenKey: []byte("my-token"),
		},
	})

	got := loadCredentialDefaultsFromClient(context.Background(), client)

	if got.Org != "my-org" {
		t.Errorf("Org = %q, want my-org", got.Org)
	}
	if got.RefreshToken != "my-token" {
		t.Errorf("RefreshToken = %q, want my-token", got.RefreshToken)
	}
}

func TestLoadCredentialDefaultsFromClientMissingOrgKey(t *testing.T) {
	client := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      credentialsSecretName,
			Namespace: credentialsNamespace,
		},
		Data: map[string][]byte{
			credentialsTokenKey: []byte("my-token"),
			// org key absent
		},
	})

	got := loadCredentialDefaultsFromClient(context.Background(), client)

	if got.Org != "" {
		t.Errorf("Org = %q, want empty", got.Org)
	}
	if got.RefreshToken != "my-token" {
		t.Errorf("RefreshToken = %q, want my-token", got.RefreshToken)
	}
}

func TestLoadCredentialDefaultsFromClientSecretNotFound(t *testing.T) {
	client := fake.NewSimpleClientset() // empty — no secret

	got := loadCredentialDefaultsFromClient(context.Background(), client)

	if got.Org != "" || got.RefreshToken != "" {
		t.Errorf("got non-empty defaults when secret missing: %+v", got)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd /home/simonc/GIT/rancher-rackspace-spot-driver && go test ./driver/ -run TestLoadCredential -v
```

Expected: compile error — `credentialsSecretName`, `credentialDefaults`, `loadCredentialDefaultsFromClient` undefined.

- [ ] **Step 3: Implement `driver/secrets.go`**

Create `driver/secrets.go`:

```go
package driver

import (
	"context"

	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	credentialsSecretName = "rackspace-spot-credentials"
	credentialsNamespace  = "cattle-system"
	credentialsOrgKey     = "org"
	credentialsTokenKey   = "refreshToken"
)

type credentialDefaults struct {
	Org          string
	RefreshToken string
}

// defaultCredentialLoader is called by stateFromOptions and GetDriverCreateOptions
// to read credential defaults. Replaced in tests to avoid real k8s access.
var defaultCredentialLoader func(ctx context.Context) credentialDefaults = loadCredentialDefaults

func loadCredentialDefaults(ctx context.Context) credentialDefaults {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		logrus.Debugf("[%s] not running in-cluster, skipping credential secret lookup: %v", driverName, err)
		return credentialDefaults{}
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logrus.Warnf("[%s] failed to create k8s client for credential lookup: %v", driverName, err)
		return credentialDefaults{}
	}
	return loadCredentialDefaultsFromClient(ctx, client)
}

func loadCredentialDefaultsFromClient(ctx context.Context, client kubernetes.Interface) credentialDefaults {
	secret, err := client.CoreV1().Secrets(credentialsNamespace).Get(ctx, credentialsSecretName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			logrus.Warnf("[%s] failed to read credential secret %s/%s: %v",
				driverName, credentialsNamespace, credentialsSecretName, err)
		}
		return credentialDefaults{}
	}
	return credentialDefaults{
		Org:          string(secret.Data[credentialsOrgKey]),
		RefreshToken: string(secret.Data[credentialsTokenKey]),
	}
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
cd /home/simonc/GIT/rancher-rackspace-spot-driver && go test ./driver/ -run TestLoadCredential -v
```

Expected:
```
--- PASS: TestLoadCredentialDefaultsFromClientBothKeys
--- PASS: TestLoadCredentialDefaultsFromClientMissingOrgKey
--- PASS: TestLoadCredentialDefaultsFromClientSecretNotFound
PASS
```

- [ ] **Step 5: Run full test suite to confirm no regressions**

```bash
cd /home/simonc/GIT/rancher-rackspace-spot-driver && go test ./...
```

Expected: all existing tests pass, 3 new tests pass.

- [ ] **Step 6: Commit**

```bash
git add driver/secrets.go driver/secrets_test.go
git commit -m "feat: add loadCredentialDefaultsFromClient for k8s Secret-based credential defaults"
```

---

## Task 2: Fallback fill in `stateFromOptions` (`driver/config.go` + `driver/config_test.go`)

**Files:**
- Modify: `driver/config.go`
- Modify: `driver/config_test.go`

- [ ] **Step 1: Write the failing tests**

Add to `driver/config_test.go` (after the last existing test):

```go
func TestStateFromOptionsFillsFromSecret(t *testing.T) {
	old := defaultCredentialLoader
	defaultCredentialLoader = func(_ context.Context) credentialDefaults {
		return credentialDefaults{Org: "secret-org", RefreshToken: "secret-token"}
	}
	defer func() { defaultCredentialLoader = old }()

	opts := &types.DriverOptions{
		StringOptions: map[string]string{
			"rackspaceSpotRegion": "us-east-1",
		},
		BoolOptions: map[string]bool{},
		IntOptions:  map[string]int64{},
	}

	state, err := stateFromOptions(context.Background(), opts)
	if err != nil {
		t.Fatalf("stateFromOptions() error = %v", err)
	}
	if state.Organization != "secret-org" {
		t.Fatalf("Organization = %q, want secret-org", state.Organization)
	}
	if state.RefreshToken != "secret-token" {
		t.Fatalf("RefreshToken = %q, want secret-token", state.RefreshToken)
	}
}

func TestStateFromOptionsExplicitValuesWinOverSecret(t *testing.T) {
	old := defaultCredentialLoader
	defaultCredentialLoader = func(_ context.Context) credentialDefaults {
		return credentialDefaults{Org: "secret-org", RefreshToken: "secret-token"}
	}
	defer func() { defaultCredentialLoader = old }()

	opts := &types.DriverOptions{
		StringOptions: map[string]string{
			"rackspaceSpotRefreshToken": "ui-token",
			"rackspaceSpotOrganization": "ui-org",
			"rackspaceSpotRegion":       "us-east-1",
		},
		BoolOptions: map[string]bool{},
		IntOptions:  map[string]int64{},
	}

	state, err := stateFromOptions(context.Background(), opts)
	if err != nil {
		t.Fatalf("stateFromOptions() error = %v", err)
	}
	if state.Organization != "ui-org" {
		t.Fatalf("Organization = %q, want ui-org", state.Organization)
	}
	if state.RefreshToken != "ui-token" {
		t.Fatalf("RefreshToken = %q, want ui-token", state.RefreshToken)
	}
}

func TestStateFromOptionsNoSecretNoUIFailsValidation(t *testing.T) {
	old := defaultCredentialLoader
	defaultCredentialLoader = func(_ context.Context) credentialDefaults {
		return credentialDefaults{}
	}
	defer func() { defaultCredentialLoader = old }()

	opts := &types.DriverOptions{
		StringOptions: map[string]string{
			"rackspaceSpotRegion": "us-east-1",
		},
		BoolOptions: map[string]bool{},
		IntOptions:  map[string]int64{},
	}

	_, err := stateFromOptions(context.Background(), opts)
	if err == nil {
		t.Fatal("stateFromOptions() should return error when org and token are missing")
	}
}
```

Also add `"context"` to the import block at the top of `driver/config_test.go`. The current import block is:

```go
import (
	"strings"
	"testing"

	"github.com/rancher/kontainer-engine/types"
)
```

Replace with:

```go
import (
	"context"
	"strings"
	"testing"

	"github.com/rancher/kontainer-engine/types"
)
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd /home/simonc/GIT/rancher-rackspace-spot-driver && go test ./driver/ -run "TestStateFromOptions" -v
```

Expected: compile error — `stateFromOptions` called with wrong number of arguments (the new tests pass `context.Background()` but the function doesn't accept it yet). The existing `TestStateFromOptionsAppliesDefaults` will also fail to compile after you update it in the next step.

- [ ] **Step 3: Update `stateFromOptions` in `driver/config.go`**

Change the function signature and add the fallback fill. The full updated function (replace lines 100–142 in `driver/config.go`):

```go
func stateFromOptions(ctx context.Context, opts *types.DriverOptions) (*clusterState, error) {
	s := &clusterState{
		RefreshToken:      getStringOption(opts, flagRefreshToken, "rackspaceSpotRefreshToken"),
		Organization:      getStringOption(opts, flagOrganization, "rackspaceSpotOrganization"),
		Region:            getStringOption(opts, flagRegion, "rackspaceSpotRegion"),
		KubernetesVersion: getStringOption(opts, flagK8sVersion, "kubernetesVersion"),
		CNI:               getStringOption(opts, flagCNI, "cni"),
		GPUEnabled:        getBoolOption(opts, flagGPUEnabled, "gpuEnabled"),
		PreemptionWebhook: getStringOption(opts, flagPreemptionWebhook, "preemptionWebhookUrl"),
		DeploymentType:    getStringOption(opts, flagDeploymentType, "deploymentType"),
		SpotPoolName:      getStringOption(opts, flagSpotPoolName, "spotNodePoolName"),
		SpotServerClass:   getStringOption(opts, flagSpotServerClass, "spotServerClass"),
		SpotBidPrice:      getStringOption(opts, flagSpotBidPrice, "spotBidPrice"),
		SpotAutoscaling:   getBoolOption(opts, flagSpotAutoscaling, "spotAutoscalingEnabled"),
		SpotMinNodes:      getIntOption(opts, flagSpotMinNodes, "spotAutoscalingMinNodes"),
		SpotMaxNodes:      getIntOption(opts, flagSpotMaxNodes, "spotAutoscalingMaxNodes"),
		OnDemandEnabled:   getBoolOption(opts, flagOnDemandEnabled, "onDemandEnabled"),
		OnDemandPoolName:  getStringOption(opts, flagOnDemandPoolName, "onDemandNodePoolName"),
		OnDemandClass:     getStringOption(opts, flagOnDemandClass, "onDemandServerClass"),
		OnDemandPrice:     getStringOption(opts, flagOnDemandPrice, "onDemandPricePerHour"),
	}

	if n, ok := lookupIntOption(opts, flagSpotNodeCount, "spotNodeCount"); ok {
		s.SpotNodeCount = int(n)
	}
	if n, ok := lookupIntOption(opts, flagOnDemandCount, "onDemandNodeCount"); ok {
		s.OnDemandCount = int(n)
	}

	if raw := getStringOption(opts, flagAdditionalSpotPools, "additionalSpotPools"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &s.AdditionalSpotPools); err != nil {
			return nil, fmt.Errorf("failed to parse additional spot pools: %w", err)
		}
	}

	if s.Organization == "" || s.RefreshToken == "" {
		defaults := defaultCredentialLoader(ctx)
		if s.Organization == "" {
			s.Organization = defaults.Org
		}
		if s.RefreshToken == "" {
			s.RefreshToken = defaults.RefreshToken
		}
	}

	applyDefaults(s)

	if err := validate(s); err != nil {
		return nil, err
	}

	return s, nil
}
```

Also add `"context"` to the import block in `driver/config.go`. Current imports start:

```go
import (
	"encoding/json"
	"fmt"
	"strings"
```

Replace with:

```go
import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
```

- [ ] **Step 4: Update the existing test call site in `driver/config_test.go`**

`TestStateFromOptionsAppliesDefaults` (line 174) currently calls `stateFromOptions(opts)`. Change it to:

```go
state, err := stateFromOptions(context.Background(), opts)
```

- [ ] **Step 5: Run tests to confirm they pass**

```bash
cd /home/simonc/GIT/rancher-rackspace-spot-driver && go test ./driver/ -run "TestStateFromOptions" -v
```

Expected:
```
--- PASS: TestStateFromOptionsAppliesDefaults
--- PASS: TestStateFromOptionsFillsFromSecret
--- PASS: TestStateFromOptionsExplicitValuesWinOverSecret
--- PASS: TestStateFromOptionsNoSecretNoUIFailsValidation
PASS
```

- [ ] **Step 6: Run full test suite**

```bash
cd /home/simonc/GIT/rancher-rackspace-spot-driver && go test ./...
```

Expected: all tests pass.

- [ ] **Step 7: Commit**

```bash
git add driver/config.go driver/config_test.go
git commit -m "feat: stateFromOptions falls back to k8s Secret for org and refresh token"
```

---

## Task 3: Pre-fill form defaults in `GetDriverCreateOptions` (`driver/driver.go` + `driver/driver_test.go`)

**Files:**
- Modify: `driver/driver.go`
- Create: `driver/driver_test.go`

- [ ] **Step 1: Write the failing tests**

Create `driver/driver_test.go`:

```go
package driver

import (
	"context"
	"testing"
)

func TestGetDriverCreateOptionsAppliesSecretDefaults(t *testing.T) {
	old := defaultCredentialLoader
	defaultCredentialLoader = func(_ context.Context) credentialDefaults {
		return credentialDefaults{Org: "preset-org", RefreshToken: "preset-token"}
	}
	defer func() { defaultCredentialLoader = old }()

	d := NewDriver()
	flags, err := d.GetDriverCreateOptions(context.Background())
	if err != nil {
		t.Fatalf("GetDriverCreateOptions() error = %v", err)
	}

	orgFlag := flags.Options[flagOrganization]
	if orgFlag.Default == nil || orgFlag.Default.DefaultString != "preset-org" {
		t.Fatalf("org Default = %v, want DefaultString=preset-org", orgFlag.Default)
	}
	tokenFlag := flags.Options[flagRefreshToken]
	if tokenFlag.Default == nil || tokenFlag.Default.DefaultString != "preset-token" {
		t.Fatalf("token Default = %v, want DefaultString=preset-token", tokenFlag.Default)
	}
}

func TestGetDriverCreateOptionsNoDefaultsWhenNoSecret(t *testing.T) {
	old := defaultCredentialLoader
	defaultCredentialLoader = func(_ context.Context) credentialDefaults {
		return credentialDefaults{}
	}
	defer func() { defaultCredentialLoader = old }()

	d := NewDriver()
	flags, err := d.GetDriverCreateOptions(context.Background())
	if err != nil {
		t.Fatalf("GetDriverCreateOptions() error = %v", err)
	}

	orgFlag := flags.Options[flagOrganization]
	if orgFlag.Default != nil {
		t.Fatalf("org Default = %v, want nil when no secret", orgFlag.Default)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd /home/simonc/GIT/rancher-rackspace-spot-driver && go test ./driver/ -run "TestGetDriverCreateOptions" -v
```

Expected: `FAIL` — `GetDriverCreateOptions` does not yet call `defaultCredentialLoader`.

- [ ] **Step 3: Update `driver/driver.go`**

**Change 1** — In `GetDriverCreateOptions`, add the defaults injection after the `flags` variable is built. Insert these lines immediately before `return flags, nil` (currently line 137):

```go
	defaults := defaultCredentialLoader(ctx)
	if defaults.Org != "" {
		flags.Options[flagOrganization].Default = &types.Default{DefaultString: defaults.Org}
	}
	if defaults.RefreshToken != "" {
		flags.Options[flagRefreshToken].Default = &types.Default{DefaultString: defaults.RefreshToken}
	}
```

**Change 2** — In `Create()`, update the `stateFromOptions` call (currently line 209) from:

```go
	s, err := stateFromOptions(opts)
```

to:

```go
	s, err := stateFromOptions(ctx, opts)
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
cd /home/simonc/GIT/rancher-rackspace-spot-driver && go test ./driver/ -run "TestGetDriverCreateOptions" -v
```

Expected:
```
--- PASS: TestGetDriverCreateOptionsAppliesSecretDefaults
--- PASS: TestGetDriverCreateOptionsNoDefaultsWhenNoSecret
PASS
```

- [ ] **Step 5: Run full test suite**

```bash
cd /home/simonc/GIT/rancher-rackspace-spot-driver && go test ./...
```

Expected: all tests pass.

- [ ] **Step 6: Commit**

```bash
git add driver/driver.go driver/driver_test.go
git commit -m "feat: GetDriverCreateOptions pre-fills org and token defaults from k8s Secret"
```

---

## Task 4: Helm chart RBAC + values documentation

**Files:**
- Modify: `deploy/charts/rackspacespot/values.yaml`
- Create: `deploy/charts/rackspacespot/templates/credentials-rbac.yaml`

- [ ] **Step 1: Update `deploy/charts/rackspacespot/values.yaml`**

Replace the entire file with:

```yaml
kontainerDriver:
  url: "https://github.com/teamzuzu/rancher-rackspace-spot-driver/releases/download/v0.1.0/rancher-rackspace-spot-driver_linux_amd64.tar.gz"
  checksum: "bb75d9788fb5501ec1b8e9e20c853cb9f7c49b22487e62dc4d45c3463d9d330c"
uiPlugin:
  endpoint: "https://teamzuzu.github.io/rancher-rackspace-spot-driver/extensions/rackspacespot/0.1.0"

# credentials configures optional pre-provisioned defaults for the Rackspace Spot org name
# and refresh token. Create a Secret with the keys below, then set secretName. The driver
# reads the Secret at runtime via Rancher's in-cluster service account; this chart creates
# the required RBAC. Leave secretName empty (the default) to disable this feature.
#
# Example secret:
#   kubectl create secret generic rackspace-spot-credentials \
#     --namespace cattle-system \
#     --from-literal=org=my-rackspace-org \
#     --from-literal=refreshToken=eyJ...
credentials:
  secretName: ""                   # name of an existing Secret; leave blank to disable
  namespace: "cattle-system"       # namespace containing the Secret
  orgKey: "org"                    # key name for the org name
  refreshTokenKey: "refreshToken"  # key name for the refresh token
```

- [ ] **Step 2: Create `deploy/charts/rackspacespot/templates/credentials-rbac.yaml`**

```yaml
{{- if .Values.credentials.secretName }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: rackspacespot-credential-reader
  namespace: {{ .Values.credentials.namespace }}
  labels:
    app.kubernetes.io/name: rackspacespot
    app.kubernetes.io/version: {{ trimPrefix "v" .Chart.AppVersion | quote }}
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    resourceNames: [{{ .Values.credentials.secretName | quote }}]
    verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: rackspacespot-credential-reader
  namespace: {{ .Values.credentials.namespace }}
  labels:
    app.kubernetes.io/name: rackspacespot
    app.kubernetes.io/version: {{ trimPrefix "v" .Chart.AppVersion | quote }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: rackspacespot-credential-reader
subjects:
  - kind: ServiceAccount
    name: rancher
    namespace: cattle-system
{{- end }}
```

- [ ] **Step 3: Lint the Helm chart**

```bash
helm lint /home/simonc/GIT/rancher-rackspace-spot-driver/deploy/charts/rackspacespot
```

Expected:
```
==> Linting deploy/charts/rackspacespot
[INFO] Chart.yaml: icon is recommended

1 chart(s) linted, 0 chart(s) failed
```

- [ ] **Step 4: Template-render with feature enabled to verify output**

```bash
helm template rackspacespot \
  /home/simonc/GIT/rancher-rackspace-spot-driver/deploy/charts/rackspacespot \
  --set credentials.secretName=rackspace-spot-credentials \
  | grep -A 30 "kind: Role"
```

Expected: a `Role` and `RoleBinding` block with `resourceNames: ["rackspace-spot-credentials"]`.

- [ ] **Step 5: Template-render with feature disabled to confirm no RBAC emitted**

```bash
helm template rackspacespot \
  /home/simonc/GIT/rancher-rackspace-spot-driver/deploy/charts/rackspacespot \
  | grep -c "kind: Role"
```

Expected: `0`

- [ ] **Step 6: Commit**

```bash
git add deploy/charts/rackspacespot/values.yaml \
        deploy/charts/rackspacespot/templates/credentials-rbac.yaml
git commit -m "feat: helm chart RBAC for credential secret defaults"
```

---

## Final verification

- [ ] **Run the full test suite one last time**

```bash
cd /home/simonc/GIT/rancher-rackspace-spot-driver && go test ./... -v 2>&1 | tail -20
```

Expected: all tests pass, zero failures.

- [ ] **Build the binary to confirm it compiles**

```bash
cd /home/simonc/GIT/rancher-rackspace-spot-driver && make build
```

Expected: `bin/kontainer-engine-driver-rackspacespot` produced with exit code 0.
