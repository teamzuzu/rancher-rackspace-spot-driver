package driver

import (
	"context"
	"strings"
	"testing"

	spotv1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
)

// mockSpotAPI records API calls in order so tests can verify operation sequencing.
type mockSpotAPI struct {
	calls     []string // e.g. "delete:pool-a", "update:pool-b"
	spotPools map[string]*spotv1.SpotNodePool
}

func (m *mockSpotAPI) record(op, name string) { m.calls = append(m.calls, op+":"+name) }

func (m *mockSpotAPI) ListSpotNodePools(_ context.Context, _, _ string) ([]*spotv1.SpotNodePool, error) {
	var out []*spotv1.SpotNodePool
	for _, p := range m.spotPools {
		cp := *p
		out = append(out, &cp)
	}
	return out, nil
}
func (m *mockSpotAPI) CreateSpotNodePool(_ context.Context, _ string, pool spotv1.SpotNodePool) error {
	m.record("create", pool.Name)
	cp := pool
	m.spotPools[pool.Name] = &cp
	return nil
}
func (m *mockSpotAPI) UpdateSpotNodePool(_ context.Context, _ string, pool spotv1.SpotNodePool) error {
	m.record("update", pool.Name)
	cp := pool
	m.spotPools[pool.Name] = &cp
	return nil
}
func (m *mockSpotAPI) DeleteSpotNodePool(_ context.Context, _, name string) error {
	m.record("delete", name)
	delete(m.spotPools, name)
	return nil
}

// Unused interface stubs.
func (m *mockSpotAPI) Authenticate(_ context.Context) (string, error)                               { return "", nil }
func (m *mockSpotAPI) GetCloudspace(_ context.Context, _, _ string) (*spotv1.CloudSpace, error)     { return nil, nil }
func (m *mockSpotAPI) CreateCloudspace(_ context.Context, _ spotv1.CloudSpace) error                { return nil }
func (m *mockSpotAPI) DeleteCloudspace(_ context.Context, _, _ string) error                        { return nil }
func (m *mockSpotAPI) GetCloudspaceConfig(_ context.Context, _, _ string) (string, error)           { return "", nil }
func (m *mockSpotAPI) GetOnDemandNodePool(_ context.Context, _, _ string) (*spotv1.OnDemandNodePool, error) {
	return nil, nil
}
func (m *mockSpotAPI) CreateOnDemandNodePool(_ context.Context, _ string, _ spotv1.OnDemandNodePool) error {
	return nil
}
func (m *mockSpotAPI) UpdateOnDemandNodePool(_ context.Context, _ string, _ spotv1.OnDemandNodePool) error {
	return nil
}
func (m *mockSpotAPI) ListOnDemandNodePools(_ context.Context, _, _ string) ([]*spotv1.OnDemandNodePool, error) {
	return nil, nil
}
func (m *mockSpotAPI) DeleteOnDemandNodePool(_ context.Context, _, _ string) error { return nil }

func indexOfCall(calls []string, prefix string) int {
	for i, c := range calls {
		if strings.HasPrefix(c, prefix) {
			return i
		}
	}
	return -1
}

// TestReconcileDeleteBeforeAutoscalingEnable covers the case Tony identified:
// an existing pool with autoscaling is removed in the same operation that
// enables autoscaling on another pool. The delete must happen first or the
// webhook will see two autoscaling pools and reject the create/update.
func TestReconcileDeleteBeforeAutoscalingEnable(t *testing.T) {
	poolA := &spotv1.SpotNodePool{Name: "pool-a"}
	poolA.Autoscaling.Enabled = true
	poolB := &spotv1.SpotNodePool{Name: "pool-b"}

	mock := &mockSpotAPI{
		spotPools: map[string]*spotv1.SpotNodePool{
			"pool-a": poolA, // autoscaling=true, being removed
			"pool-b": poolB, // gaining autoscaling
		},
	}

	client := &spotClient{api: mock, org: "org"}
	state := &clusterState{
		CloudspaceName:  "cs",
		SpotPoolName:    "pool-b",
		SpotServerClass: "rxtx.4xlarge-mi300x",
		SpotNodeCount:   3,
		SpotBidPrice:    "0.01",
		SpotAutoscaling: true,
		// pool-a absent from desired → deleted
	}

	if err := client.reconcileSpotNodePools(context.Background(), state); err != nil {
		t.Fatalf("reconcileSpotNodePools() error = %v", err)
	}

	delIdx := indexOfCall(mock.calls, "delete:pool-a")
	enaIdx := indexOfCall(mock.calls, "update:pool-b")
	if delIdx == -1 {
		t.Fatal("pool-a was not deleted")
	}
	if enaIdx == -1 {
		t.Fatal("pool-b was not updated")
	}
	if delIdx > enaIdx {
		t.Errorf("pool-a delete (call %d) happened after pool-b autoscaling enable (call %d); webhook would reject",
			delIdx, enaIdx)
	}
}

// TestReconcileAutoscalingSwapDisableBeforeEnable covers the case where one
// existing pool loses autoscaling while another gains it. The disable must
// run before the enable.
func TestReconcileAutoscalingSwapDisableBeforeEnable(t *testing.T) {
	poolA := &spotv1.SpotNodePool{Name: "pool-a"}
	poolA.Autoscaling.Enabled = true // losing autoscaling
	poolB := &spotv1.SpotNodePool{Name: "pool-b"}

	mock := &mockSpotAPI{
		spotPools: map[string]*spotv1.SpotNodePool{
			"pool-a": poolA,
			"pool-b": poolB,
		},
	}

	client := &spotClient{api: mock, org: "org"}
	state := &clusterState{
		CloudspaceName:  "cs",
		SpotPoolName:    "pool-a",
		SpotServerClass: "rxtx.4xlarge-mi300x",
		SpotNodeCount:   3,
		SpotBidPrice:    "0.01",
		SpotAutoscaling: false, // pool-a: autoscaling off
		AdditionalSpotPools: []SpotPoolConfig{
			{
				Name:        "pool-b",
				ServerClass: "rxtx.4xlarge-mi300x",
				NodeCount:   2,
				BidPrice:    "0.01",
				Autoscaling: true, // pool-b: autoscaling on
			},
		},
	}

	if err := client.reconcileSpotNodePools(context.Background(), state); err != nil {
		t.Fatalf("reconcileSpotNodePools() error = %v", err)
	}

	disIdx := indexOfCall(mock.calls, "update:pool-a") // phase 2: disable
	enaIdx := indexOfCall(mock.calls, "update:pool-b") // phase 3: enable
	if disIdx == -1 {
		t.Fatal("pool-a autoscaling disable not recorded")
	}
	if enaIdx == -1 {
		t.Fatal("pool-b autoscaling enable not recorded")
	}
	if disIdx > enaIdx {
		t.Errorf("pool-a disable (call %d) happened after pool-b enable (call %d); webhook would reject",
			disIdx, enaIdx)
	}
}
