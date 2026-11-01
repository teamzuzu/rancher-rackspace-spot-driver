package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	spotv1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/rancher/kontainer-engine/types"
)

const autopilotTestName = "11111111-1111-4111-8111-111111111111"

func validAutopilotPool() AutopilotPoolConfig {
	return AutopilotPoolConfig{Name: autopilotTestName, VCPUs: 16, MemoryPerVCPU: "4", BudgetPerHour: "0.50", MinVCPUsPerNode: 2, MaxVCPUsPerNode: 8}
}

func TestAutopilotOptionsAndState(t *testing.T) {
	for _, key := range []string{flagAutopilotPools, "autopilotPools"} {
		t.Run(key, func(t *testing.T) {
			opts := &types.DriverOptions{StringOptions: map[string]string{flagRefreshToken: "secret", flagOrganization: "org", flagRegion: "us-east-iad-1", key: `[{"vcpus":16,"budgetPerHour":"0.50"}]`}, BoolOptions: map[string]bool{flagDisablePrimarySpotPool: true}}
			s, err := stateFromOptions(opts)
			if err != nil {
				t.Fatal(err)
			}
			if len(s.AutopilotPools) != 1 || s.AutopilotPools[0].Name == "" || s.AutopilotPools[0].MemoryPerVCPU != "any" || !s.DisablePrimarySpotPool {
				t.Fatalf("unexpected state: %+v", s)
			}
			info := &types.ClusterInfo{}
			if err := s.save(info); err != nil {
				t.Fatal(err)
			}
			restored, err := stateFromClusterInfo(info)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(s, restored) {
				t.Fatalf("state did not round trip")
			}
			if strings.Contains(info.Metadata[metaStateKey], "secret") {
				t.Fatal("token leaked into metadata")
			}
		})
	}
	s, err := stateFromClusterInfo(&types.ClusterInfo{Metadata: map[string]string{metaStateKey: `{"spotPoolName":"existing"}`}})
	if err != nil || s.DisablePrimarySpotPool || len(s.AutopilotPools) != 0 {
		t.Fatalf("legacy state incompatible: %v", err)
	}
}

func TestAutopilotValidation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*AutopilotPoolConfig)
	}{
		{"name", func(p *AutopilotPoolConfig) { p.Name = "invalid" }},
		{"capacity", func(p *AutopilotPoolConfig) { p.VCPUs = 0 }},
		{"memory", func(p *AutopilotPoolConfig) { p.MemoryPerVCPU = "3" }},
		{"budget", func(p *AutopilotPoolConfig) { p.BudgetPerHour = "NaN" }},
		{"negative budget", func(p *AutopilotPoolConfig) { p.BudgetPerHour = "-0.1" }},
		{"zero budget", func(p *AutopilotPoolConfig) { p.BudgetPerHour = "0" }},
		{"bounds", func(p *AutopilotPoolConfig) { p.MinVCPUsPerNode = 10 }},
		{"negative bounds", func(p *AutopilotPoolConfig) { p.MinVCPUsPerNode = -1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := validAutopilotPool()
			tc.change(&p)
			if validateAutopilotPools([]AutopilotPoolConfig{p}) == nil {
				t.Fatal("invalid pool accepted")
			}
		})
	}
	if err := validateAutopilotPools([]AutopilotPoolConfig{validAutopilotPool()}); err != nil {
		t.Fatal(err)
	}
	if validateAutopilotPools([]AutopilotPoolConfig{validAutopilotPool(), validAutopilotPool()}) == nil {
		t.Fatal("duplicate names accepted")
	}
}

func newTestAutopilotClient(t *testing.T, handler http.HandlerFunc) *autopilotClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	// A long-lived synthetic JWT keeps SDK organization lookups local.
	return &autopilotClient{&spotv1.RackspaceSpotClient{BaseURL: server.URL, HTTPClient: server.Client(), Token: "header.eyJleHAiOjQxMDI0NDQ4MDB9.signature"}}
}

func TestAutopilotHTTPContract(t *testing.T) {
	var calls []string
	var bodies []map[string]any
	c := newTestAutopilotClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("missing authentication")
		}
		if r.URL.Path == "/apis/auth.ngpc.rxt.io/v1/organizations" {
			fmt.Fprint(w, `{"organizations":[{"id":"org_test","name":"test"}]}`)
			return
		}
		base := "/apis/ngpc.rxt.io/v1/namespaces/org-test/autopilotnodepools"
		if r.URL.Path != base && r.URL.Path != base+"/"+autopilotTestName {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		calls = append(calls, r.Method)
		if r.Method == http.MethodGet {
			if r.URL.Query().Get("labelSelector") != "ngpc.rxt.io/cloudspace=cs" {
				t.Error("missing cloudspace scope")
			}
			fmt.Fprintf(w, `{"items":[{"metadata":{"name":%q},"spec":{"vcpu":{"total":16},"memoryPerVCPU":"4","budgetPerHour":"0.50","vcpuPerNode":{"min":2,"max":8}}}]}`, autopilotTestName)
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPatch {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			bodies = append(bodies, body)
			if r.Method == http.MethodPatch && r.Header.Get("Content-Type") != "application/merge-patch+json" {
				t.Error("wrong patch media type")
			}
		}
		w.WriteHeader(http.StatusOK)
	})
	ctx := context.Background()
	p := validAutopilotPool()
	if err := c.WriteAutopilotPool(ctx, "test", "cs", "us-east-iad-1", p, false); err != nil {
		t.Fatal(err)
	}
	got, err := c.ListAutopilotPools(ctx, "test", "cs")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []AutopilotPoolConfig{p}) {
		t.Fatalf("decoded pools: %+v", got)
	}
	p.MinVCPUsPerNode = 0
	p.MaxVCPUsPerNode = 0
	if err := c.WriteAutopilotPool(ctx, "test", "cs", "us-east-iad-1", p, true); err != nil {
		t.Fatal(err)
	}
	p.MaxVCPUsPerNode = 8
	if err := c.WriteAutopilotPool(ctx, "test", "cs", "us-east-iad-1", p, true); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteAutopilotPool(ctx, "test", p.Name); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(calls, []string{"POST", "GET", "PATCH", "PATCH", "DELETE"}) {
		t.Fatal(calls)
	}
	create := bodies[0]
	spec := create["spec"].(map[string]any)
	if create["kind"] != "AutopilotNodePool" || spec["allocationStrategy"] != "LowestPrice" || spec["cloudSpace"] != "cs" || spec["region"] != "us-east-iad-1" {
		t.Fatalf("bad payload: %+v", create)
	}
	if bodies[1]["spec"].(map[string]any)["vcpuPerNode"] != nil {
		t.Fatal("unset bounds not cleared")
	}
	bounds := bodies[2]["spec"].(map[string]any)["vcpuPerNode"].(map[string]any)
	if bounds["min"] != nil || bounds["max"] != float64(8) {
		t.Fatal("partial bounds not cleared")
	}
}

func TestAutopilotGeneratedPoolsExcluded(t *testing.T) {
	c := newTestAutopilotClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/organizations") {
			fmt.Fprint(w, `{"organizations":[{"id":"org_test","name":"test"}]}`)
			return
		}
		fmt.Fprint(w, `{"items":[{"metadata":{"name":"manual"},"spec":{"desired":3,"autoscaling":{"enabled":true,"minNodes":1,"maxNodes":5}}},{"metadata":{"name":"generated","labels":{"autopilotnodepool.ngpc.rxt.io/name":"parent"}},"spec":{"desired":4}}]}`)
	})
	pools, err := c.ListSpotNodePools(context.Background(), "test", "cs")
	if err != nil {
		t.Fatal(err)
	}
	if len(pools) != 1 || pools[0].Name != "manual" || !pools[0].Autoscaling.Enabled {
		t.Fatalf("incorrect manual pools: %+v", pools)
	}
	original := &spotv1.CloudSpace{SpotNodepools: []*spotv1.SpotNodePool{{Name: "generated"}, {Name: "manual"}}}
	cs, err := c.manualCloudspace(context.Background(), "test", "cs", original)
	if err != nil {
		t.Fatal(err)
	}
	if len(cs.SpotNodepools) != 1 || len(original.SpotNodepools) != 2 {
		t.Fatal("import changed original or included generated pools")
	}
}

type mockAutopilotAPI struct {
	pools []AutopilotPoolConfig
	calls []string
	err   error
}

func (m *mockAutopilotAPI) ListAutopilotPools(context.Context, string, string) ([]AutopilotPoolConfig, error) {
	return m.pools, m.err
}
func (m *mockAutopilotAPI) WriteAutopilotPool(_ context.Context, _, _, _ string, p AutopilotPoolConfig, exists bool) error {
	m.calls = append(m.calls, fmt.Sprintf("write:%s:%t", p.Name, exists))
	return m.err
}
func (m *mockAutopilotAPI) DeleteAutopilotPool(_ context.Context, _ string, name string) error {
	m.calls = append(m.calls, "delete:"+name)
	return m.err
}

func TestAutopilotReconcileAndImport(t *testing.T) {
	p := validAutopilotPool()
	mock := &mockAutopilotAPI{pools: []AutopilotPoolConfig{p, {Name: "removed"}}}
	client := &spotClient{autopilot: mock, org: "test"}
	state := &clusterState{CloudspaceName: "cs", AutopilotPools: []AutopilotPoolConfig{p}}
	if err := client.reconcileAutopilotNodePools(context.Background(), state); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(mock.calls, []string{"write:" + p.Name + ":true", "delete:removed"}) {
		t.Fatal(mock.calls)
	}
	mock.calls = nil
	mock.pools = []AutopilotPoolConfig{p}
	state.AutopilotPools = nil
	if err := client.reconcileAutopilotNodePools(context.Background(), state); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(mock.calls, []string{"delete:" + p.Name}) {
		t.Fatal(mock.calls)
	}
	mock.calls = nil
	mock.pools = nil
	state.AutopilotPools = []AutopilotPoolConfig{p}
	if err := client.reconcileAutopilotNodePools(context.Background(), state); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(mock.calls, []string{"write:" + p.Name + ":false"}) {
		t.Fatal(mock.calls)
	}
	mock.calls = nil
	mock.pools = []AutopilotPoolConfig{p}
	client.api = &importMockAPI{cloudspace: &spotv1.CloudSpace{Name: "cs"}}
	imported, err := client.importCloudspace(context.Background(), "test", "cs", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(imported.AutopilotPools, mock.pools) || !imported.DisablePrimarySpotPool || !imported.Imported || len(mock.calls) > 0 {
		t.Fatalf("bad import: %+v", imported)
	}
}

func TestAutopilotOnlySkipsPrimaryManualPool(t *testing.T) {
	mock := &mockSpotAPI{spotPools: map[string]*spotv1.SpotNodePool{}}
	c := &spotClient{api: mock}
	state := &clusterState{DisablePrimarySpotPool: true, SpotAutoscaling: true}
	applyDefaults(state)
	if err := c.reconcileSpotNodePools(context.Background(), state); err != nil {
		t.Fatal(err)
	}
	if len(mock.calls) > 0 {
		t.Fatal("created unwanted manual pool", mock.calls)
	}
}

func TestAutopilotUpdateRejectsInvalidJSONBeforeAPI(t *testing.T) {
	state := &clusterState{RefreshToken: "secret", Organization: "test", Region: "us-east-iad-1", AutopilotPools: []AutopilotPoolConfig{validAutopilotPool()}}
	info := &types.ClusterInfo{}
	if err := state.save(info); err != nil {
		t.Fatal(err)
	}
	before := info.Metadata[metaStateKey]
	_, err := (&Driver{}).Update(context.Background(), info, &types.DriverOptions{StringOptions: map[string]string{"autopilotPools": "invalid"}})
	if err == nil || !strings.Contains(err.Error(), "invalid autopilot pools") {
		t.Fatalf("got %v", err)
	}
	if info.Metadata[metaStateKey] != before {
		t.Fatal("invalid update changed stored state")
	}
}

func TestAutopilotFailuresStopReconciliation(t *testing.T) {
	mock := &mockAutopilotAPI{err: fmt.Errorf("access denied")}
	client := &spotClient{autopilot: mock}
	if err := client.reconcileAutopilotNodePools(context.Background(), &clusterState{AutopilotPools: []AutopilotPoolConfig{validAutopilotPool()}}); err == nil {
		t.Fatal("list error ignored")
	}
	if len(mock.calls) > 0 {
		t.Fatal("mutations after list failure")
	}
	c := newTestAutopilotClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "sensitive response")
	})
	err := c.request(context.Background(), http.MethodGet, "/test", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "sensitive") {
		t.Fatalf("bad HTTP error: %v", err)
	}
}

func TestAutopilotRemovalAndImportedClusterProtection(t *testing.T) {
	mock := &mockAutopilotAPI{pools: []AutopilotPoolConfig{validAutopilotPool()}}
	api := &mockSpotAPI{spotPools: map[string]*spotv1.SpotNodePool{}}
	client := &spotClient{autopilot: mock, api: api}
	if err := client.deleteNodePools(context.Background(), &clusterState{CloudspaceName: "cs"}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(mock.calls, []string{"delete:" + autopilotTestName}) {
		t.Fatal(mock.calls)
	}
	info := &types.ClusterInfo{}
	s := &clusterState{Imported: true, CloudspaceName: "cs", AutopilotPools: mock.pools}
	if err := s.save(info); err != nil {
		t.Fatal(err)
	}
	if err := (&Driver{}).Remove(context.Background(), info); err != nil {
		t.Fatal(err)
	}
}

func TestAutopilotFlagsAndNodeCountScalingGuard(t *testing.T) {
	d := &Driver{}
	for _, getFlags := range []func(context.Context) (*types.DriverFlags, error){d.GetDriverCreateOptions, d.GetDriverUpdateOptions} {
		flags, err := getFlags(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if flags.Options[flagAutopilotPools] == nil || flags.Options[flagDisablePrimarySpotPool] == nil {
			t.Fatal("missing Autopilot flags")
		}
	}
	info := &types.ClusterInfo{}
	s := &clusterState{DisablePrimarySpotPool: true}
	if err := s.save(info); err != nil {
		t.Fatal(err)
	}
	if err := d.SetClusterSize(context.Background(), info, &types.NodeCount{Count: 3}); err == nil || !strings.Contains(err.Error(), "vCPU") {
		t.Fatalf("expected scaling guard, got %v", err)
	}
}
