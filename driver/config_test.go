package driver

import (
	"strings"
	"testing"

	"github.com/rancher/kontainer-engine/types"
)

func TestSaveRedactsRefreshTokenFromMetadata(t *testing.T) {
	state := &clusterState{
		RefreshToken:      "refresh-token",
		Organization:      "org",
		CloudspaceName:    "cluster",
		Region:            defaultRegion,
		KubernetesVersion: defaultK8sVersion,
		CNI:               defaultCNI,
		SpotPoolName:      "pool",
		SpotServerClass:   defaultSpotClass,
		SpotNodeCount:     int(defaultSpotCount),
		SpotBidPrice:      defaultSpotBid,
	}
	info := &types.ClusterInfo{}

	if err := state.save(info); err != nil {
		t.Fatalf("save() error = %v", err)
	}

	if info.Password != "refresh-token" {
		t.Fatalf("Password = %q, want refresh token", info.Password)
	}
	if strings.Contains(info.Metadata[metaStateKey], "refresh-token") {
		t.Fatalf("metadata contains refresh token: %s", info.Metadata[metaStateKey])
	}
}

func TestStateFromClusterInfoSupportsLegacyMetadataToken(t *testing.T) {
	info := &types.ClusterInfo{
		Metadata: map[string]string{
			metaStateKey: `{"refreshToken":"legacy-token","organization":"org"}`,
		},
	}

	state, err := stateFromClusterInfo(info)
	if err != nil {
		t.Fatalf("stateFromClusterInfo() error = %v", err)
	}

	if state.RefreshToken != "legacy-token" {
		t.Fatalf("RefreshToken = %q, want legacy-token", state.RefreshToken)
	}
}

func TestStateFromClusterInfoReadsTokenFromPassword(t *testing.T) {
	info := &types.ClusterInfo{
		Password: "stored-token",
		Metadata: map[string]string{
			metaStateKey: `{"organization":"org"}`,
		},
	}

	state, err := stateFromClusterInfo(info)
	if err != nil {
		t.Fatalf("stateFromClusterInfo() error = %v", err)
	}

	if state.RefreshToken != "stored-token" {
		t.Fatalf("RefreshToken = %q, want stored-token", state.RefreshToken)
	}
}

func TestMergeStateAppliesFalseAndZeroValues(t *testing.T) {
	state := &clusterState{
		RefreshToken:        "old-token",
		SpotAutoscaling:     true,
		SpotNodeCount:       3,
		SpotMinNodes:        1,
		SpotMaxNodes:        10,
		OnDemandEnabled:     true,
		OnDemandCount:       2,
		AdditionalSpotPools: []SpotPoolConfig{{Name: "old"}},
	}
	opts := &types.DriverOptions{
		BoolOptions: map[string]bool{
			"spotAutoscalingEnabled": false,
			"onDemandEnabled":        false,
		},
		IntOptions: map[string]int64{
			"spotNodeCount":           0,
			"spotAutoscalingMinNodes": 0,
			"spotAutoscalingMaxNodes": 0,
			"onDemandNodeCount":       0,
		},
		StringOptions: map[string]string{
			"rackspaceSpotRefreshToken": "new-token",
			"additionalSpotPools":       "[]",
		},
	}

	mergeState(state, opts)

	if state.RefreshToken != "new-token" {
		t.Fatalf("RefreshToken = %q, want new-token", state.RefreshToken)
	}
	if state.SpotAutoscaling {
		t.Fatal("SpotAutoscaling = true, want false")
	}
	if state.OnDemandEnabled {
		t.Fatal("OnDemandEnabled = true, want false")
	}
	if state.SpotNodeCount != 0 || state.SpotMinNodes != 0 || state.SpotMaxNodes != 0 || state.OnDemandCount != 0 {
		t.Fatalf("zero values not applied: %+v", state)
	}
	if len(state.AdditionalSpotPools) != 0 {
		t.Fatalf("AdditionalSpotPools len = %d, want 0", len(state.AdditionalSpotPools))
	}
}

func TestStateFromOptionsDefaultsAndRedaction(t *testing.T) {
	opts := &types.DriverOptions{
		StringOptions: map[string]string{
			"rackspaceSpotRefreshToken": "token",
			"rackspaceSpotOrganization": "org",
		},
		BoolOptions: map[string]bool{},
		IntOptions:  map[string]int64{},
	}

	state, err := stateFromOptions(opts)
	if err != nil {
		t.Fatalf("stateFromOptions() error = %v", err)
	}

	if state.Region != defaultRegion || state.KubernetesVersion != defaultK8sVersion || state.SpotServerClass != defaultSpotClass {
		t.Fatalf("defaults not applied: %+v", state)
	}
}
